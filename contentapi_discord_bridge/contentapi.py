import asyncio
from dataclasses import dataclass
from enum import Enum
import itertools
import json
import signal
import traceback
import websockets
from websockets.asyncio.client import connect, ClientConnection
from typing import Any, Awaitable, Callable

import aiohttp
from sqlalchemy import event


@dataclass
class User:
    id: int
    name: str
    avatar: str


class MessageEventType(Enum):
    CREATED = 1
    UPDATED = 2
    DELETED = 3


@dataclass
class Message:
    id: int
    text: str
    markup: str


@dataclass
class MessageEvent:
    message: Message
    state: MessageEventType
    user: User | None
    content_id: int


type MessageFunc = Callable[[MessageEvent], Awaitable[None]]


async def keepalive(websocket: ClientConnection, ping_interval: int = 30):
    while True:
        await asyncio.sleep(ping_interval)
        try:
            await websocket.send(json.dumps({"type": "ping"}))
        except websockets.ConnectionClosed:
            break


class ContentApi:
    def __init__(self, domain: str, token: str):
        self.domain: str = domain
        self.token: str = token
        self._uid = None
        self._on_created_listeners: list[MessageFunc] = []
        self._on_updated_listeners: list[MessageFunc] = []
        self._on_deleted_listeners: list[MessageFunc] = []

    def api_route(self) -> str:
        return f"https://{self.domain}/api"

    def file_route(self, file_hash: str) -> str:
        return f"{self.api_route()}/File/raw/{file_hash}"

    def get_avatar(self, user: User, size: int = 100) -> str:
        return f"{self.file_route(user.avatar)}?size={size}&crop=true"

    def _authorized_headers(self) -> dict[str, str]:
        return {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.token}",
        }

    def _authorized_blank_headers(self) -> dict[str, str]:
        return {
            "Authorization": f"Bearer {self.token}",
        }

    async def _post_message(
        self, message: dict[str, str | int | dict[str, str]]
    ) -> int:
        async with aiohttp.ClientSession() as session:
            json_data = json.dumps(message)
            async with session.post(
                f"{self.api_route()}/Write/message",
                headers=self._authorized_headers(),
                data=json_data,
            ) as resp:
                data = await resp.json()
                return int(data["id"])

    async def write_message(
        self, room_id: int, content: str, username: str, avatar: str, markup: str
    ) -> int:
        return await self._post_message(
            {
                "text": content,
                "contentid": room_id,
                "values": {
                    # nickname
                    "n": username,
                    # markup
                    "m": markup,
                    # avatar
                    "a": avatar,
                },
            }
        )

    async def edit_message(
        self,
        msg_id: int,
        room_id: int,
        content: str,
        username: str,
        avatar: str,
        markup: str,
    ) -> int:
        return await self._post_message(
            {
                "id": msg_id,
                "text": content,
                "contentid": room_id,
                "values": {
                    # nickname
                    "n": username,
                    # markup
                    "m": markup,
                    # avatar
                    "a": avatar,
                },
            }
        )

    async def delete_message(self, msg_id: int) -> None:
        async with aiohttp.ClientSession(
            headers=self._authorized_blank_headers()
        ) as session:
            async with session.post(
                f"{self.api_route()}/Delete/message/{msg_id}",
            ) as resp:
                print(await resp.text())
                return

    async def upload_file(
        self, file_name: str, file: bytes, bucket: str | None = None
    ) -> str:
        async with aiohttp.ClientSession(
            headers=self._authorized_blank_headers()
        ) as session:
            data = aiohttp.FormData()
            data.add_field("file", file, filename=file_name)
            if bucket is not None:
                data.add_field("globalPerms", ".")
                data.add_field("values[bucket]", bucket)

            async with session.post(f"{self.api_route()}/File", data=data) as resp:
                data = await resp.json()
                return str(data["hash"])

    async def load_user_id(self):
        _ = await self.get_user_id()

    async def get_user_id(self) -> int:
        async with aiohttp.ClientSession() as session:
            async with session.get(f"{self.api_route()}/User/me") as resp:
                data = await resp.json()
                self._uid = int(data["id"])
                return self._uid

    @property
    def uid(self):
        if self._uid is None:
            raise ValueError("User ID is not loaded")
        return self._uid

    @uid.getter
    def uid(self):
        return self._uid

    async def get_room_name(self, room_id: int) -> str:
        async with aiohttp.ClientSession(headers=self._authorized_headers()) as session:
            async with session.post(
                f"{self.api_route()}/Request",
                data=json.dumps(
                    {
                        "values": {"key": room_id},
                        "requests": [
                            {
                                "type": "content",
                                "fields": "id,name",
                                "query": "id = @key",
                            }
                        ],
                    }
                ),
            ) as resp:
                data = await resp.json()
                print(data)
                return str(data["objects"]["content"][0]["name"])

    def on_message_created(self, func: MessageFunc) -> None:
        self._on_created_listeners.append(func)

    def on_message_updated(self, func: MessageFunc) -> None:
        self._on_updated_listeners.append(func)

    def on_message_deleted(self, func: MessageFunc) -> None:
        self._on_deleted_listeners.append(func)

    @staticmethod
    def parse_message_events(input: str) -> list[MessageEvent]:
        events: list[MessageEvent] = []
        data: dict[Any, Any] = json.loads(input)

        if data["type"] == "live":
            raw_events = data["data"]["events"]
            raw_objects = data["data"]["objects"]

            if "message_event" not in raw_objects:
                return []

            message_event_list = raw_objects["message_event"]
            if not message_event_list:
                raise ValueError("message_event object not found")

            message_list = message_event_list["message"]
            if not message_list:
                raise ValueError("message object not found")

            user_data_list = message_event_list["user"]
            if not user_data_list:
                raise ValueError("user list object not found")

            for raw_event in raw_events:
                ref_id = raw_event["refId"]

                raw_message = next(
                    (
                        raw_message
                        for raw_message in message_list
                        if raw_message["id"] == ref_id
                    ),
                    None,
                )

                if not raw_message:
                    continue

                matching_user = next(
                    (
                        user
                        for user in user_data_list
                        if user["id"] == raw_message["createUserId"]
                    ),
                    None,
                )

                values = raw_message["values"]

                avatar = "5413"
                if "a" in values:
                    avatar = values["a"]
                elif matching_user is not None:
                    avatar = matching_user["avatar"]

                user = (
                    User(
                        id=matching_user["id"],
                        name=matching_user["username"],
                        avatar=avatar,
                    )
                    if matching_user
                    else None
                )

                message = Message(
                    id=ref_id,
                    text=raw_message["text"],
                    markup="plaintext",
                )

                if "m" in values:
                    message.markup = values["m"]

                state = MessageEventType.CREATED
                if int(raw_message["deleted"]) == 1:
                    state = MessageEventType.DELETED
                elif int(raw_message["edited"]) == 1:
                    state = MessageEventType.UPDATED

                events.append(
                    MessageEvent(
                        message=message,
                        state=state,
                        user=user,
                        content_id=raw_message["contentId"],
                    )
                )

        return events

    async def socket(self):
        await self.load_user_id()
        while True:
            async with connect(
                f"wss://{self.domain}/api/live/ws?token={self.token}"
            ) as ws:
                keepalive_task = asyncio.create_task(keepalive(ws))
                try:
                    async for msg in ws:
                        events = self.parse_message_events(str(msg))
                        for event in events:
                            if event.state == MessageEventType.CREATED:
                                for i in self._on_created_listeners:
                                    await i(event)
                            elif event.state == MessageEventType.UPDATED:
                                for i in self._on_updated_listeners:
                                    await i(event)
                            elif event.state == MessageEventType.DELETED:
                                for i in self._on_deleted_listeners:
                                    await i(event)
                except websockets.ConnectionClosed:
                    print("Connection closed")
                except Exception:
                    print("Unexpected error")
                    print(traceback.format_exc())
                finally:
                    _ = keepalive_task.cancel()
                    print("Will attempt to reconnect in 15 seconds")
                    await asyncio.sleep(15)
