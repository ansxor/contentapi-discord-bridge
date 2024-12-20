#!/usr/bin/env python

import asyncio
import os
import tempfile
import traceback
from typing import override
import aiohttp
import nextcord
from nextcord.ext.commands import Bot
from nextcord.state import TextChannel
import sqlalchemy as sqla
from sqlalchemy.orm import DeclarativeBase
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker
from .contentapi import ContentApi
from .markup import MarkupService
from contentapi_discord_bridge import contentapi

content_api = ContentApi(
    os.environ["CONTENTAPI_DOMAIN"],
    os.environ["CONTENTAPI_TOKEN"],
)
markup_service = MarkupService(os.environ["MARKUP_SERVICE_DOMAIN"])


class Base(DeclarativeBase):
    pass


class ChannelPair(Base):
    __tablename__: str = "channel_store"

    discord_channel_id: sqla.Column[int] = sqla.Column(
        sqla.Integer, primary_key=True, nullable=False
    )
    content_api_room_id: sqla.Column[int] = sqla.Column(sqla.Integer, nullable=False)

    @override
    def __repr__(self) -> str:
        return f"<ChannelPair(discord_channel_id={self.discord_channel_id}, content_api_room_id={self.content_api_room_id})>"


class WebhookChannelPair(Base):
    __tablename__: str = "webhook_channel_store"

    discord_channel_id: sqla.Column[int] = sqla.Column(
        sqla.Integer, primary_key=True, nullable=False
    )
    webhook_url: sqla.Column[str] = sqla.Column(sqla.Text, nullable=False)

    @override
    def __repr__(self) -> str:
        return f"<WebhookChannelPair(discord_channel_id={self.discord_channel_id}, webhook_url={self.webhook_url})>"


class DiscordAttachment(Base):
    __tablename__: str = "discord_attachment_store"

    attachment_url: sqla.Column[str] = sqla.Column(sqla.Text, primary_key=True)
    content_api_hash: sqla.Column[str] = sqla.Column(sqla.Text, nullable=False)

    @override
    def __repr__(self) -> str:
        return f"<DiscordAttachment(attachment_url={self.attachment_url}, content_api_hash={self.content_api_hash})>"

    @staticmethod
    async def get_attachment(attachment: nextcord.Attachment) -> str:
        async with async_session.begin() as session:
            url = attachment.url.split("?")[0]
            existing_pair = await session.get(DiscordAttachment, url)
            if existing_pair is not None:
                return str(existing_pair.content_api_hash)

            with tempfile.TemporaryFile() as temp_file:
                _ = await attachment.save(temp_file)
                _ = temp_file.seek(0)
                bytes = temp_file.read()

                hash = await content_api.upload_file(
                    attachment.filename,
                    bytes,
                    "discord-bridge-upload",
                )

                session.add(
                    DiscordAttachment(
                        attachment_url=url,
                        content_api_hash=hash,
                    )
                )
                await session.commit()

                return hash


class AvatarStore(Base):
    __tablename__: str = "avatar_store"

    discord_uid: sqla.Column[str] = sqla.Column(
        sqla.Text, primary_key=True, nullable=False
    )
    discord_avatar_url: sqla.Column[str] = sqla.Column(
        sqla.Text, nullable=False, unique=True
    )
    content_api_hash: sqla.Column[str] = sqla.Column(
        sqla.Text, nullable=False, unique=True
    )

    @override
    def __repr__(self) -> str:
        return f"<AvatarStore(discord_user_id={self.discord_uid}, discord_avatar_url={self.discord_avatar_url}, content_api_user_id={self.content_api_hash})>"

    @staticmethod
    async def fetch_avatar_from_user(
        user: nextcord.User | nextcord.Member, content_api: ContentApi
    ) -> str:
        async with async_session.begin() as session:
            avatar = user.avatar
            if avatar is None:
                avatar = user.default_avatar

            existing_pair = await session.get(AvatarStore, user.id)
            if existing_pair is not None:
                if str(existing_pair.discord_avatar_url) == str(avatar.url):
                    return str(existing_pair.content_api_hash)
                else:
                    await session.delete(existing_pair)
                    session.commit()

        async with async_session.begin() as session:
            with tempfile.TemporaryFile() as temp_file:
                _ = await avatar.save(temp_file)
                _ = temp_file.seek(0)
                bytes = temp_file.read()

                hash = await content_api.upload_file(
                    "avatar.webp",
                    bytes,
                    "discord-bridge-avatars",
                )

                avatar_pair = AvatarStore(
                    discord_uid=user.id,
                    discord_avatar_url=avatar.url,
                    content_api_hash=hash,
                )
                session.add(avatar_pair)
                await session.commit()

                return hash


class ContentApiMessageStore(Base):
    __tablename__: str = "content_api_message_store"

    discord_message_id: sqla.Column[str] = sqla.Column(
        sqla.Text, nullable=False, primary_key=True
    )
    content_api_message_id: sqla.Column[int] = sqla.Column(
        sqla.Integer, nullable=False, unique=True
    )
    content_api_room_id: sqla.Column[int] = sqla.Column(sqla.Integer, nullable=False)

    @override
    def __repr__(self) -> str:
        return f"<ContentApiMessageStore(content_api_message_id={self.content_api_message_id}, content_api_room_id={self.content_api_room_id}, discord_message_id={self.discord_message_id})>"


class WebhookMessageStore(Base):
    __tablename__: str = "webhook_message_store"

    discord_message_id: sqla.Column[int] = sqla.Column(
        sqla.Integer, nullable=False, primary_key=True
    )
    webhook_id: sqla.Column[int] = sqla.Column(sqla.Integer, nullable=False)
    webhook_message_channel_id: sqla.Column[int] = sqla.Column(
        sqla.Integer, nullable=False
    )
    contentapi_message_id: sqla.Column[int] = sqla.Column(sqla.Integer, nullable=False)

    @override
    def __repr__(self) -> str:
        return f"<WebhookMessageStore(discord_message_id={self.discord_message_id}, webhook_id={self.webhook_id}, webhook_message_channel_id={self.webhook_message_channel_id}, contentapi_message_id={self.contentapi_message_id})>"


engine = create_async_engine(f"sqlite+aiosqlite:///{os.environ['DB_FILE']}")
async_session = async_sessionmaker(engine, expire_on_commit=False)


intents = nextcord.Intents.default()
intents.message_content = True
bot = Bot(intents=intents)


@bot.slash_command(force_global=True)
async def bind(interaction: nextcord.Interaction[nextcord.Client], room_id: int):
    """Binds a Discord channel to a Content API channel.

    Parameters
    ----------
    interaction : nextcord.Interaction[nextcord.Client]
        The interaction that triggered this command.
    room_id : int
        The ID of the Content API room to bind.
    """
    channel_id = interaction.channel_id
    if channel_id is None:
        _ = await interaction.response.send_message(
            "This command must be run in a Discord channel."
        )
        return
    channel_id = str(channel_id)

    try:
        room_name = await content_api.get_room_name(room_id)

        async with async_session.begin() as session:
            old_channel_pair = await session.get(ChannelPair, channel_id)
            if old_channel_pair is not None:
                await session.delete(old_channel_pair)
                await session.commit()

        async with async_session.begin() as session:
            session.add(
                ChannelPair(discord_channel_id=channel_id, content_api_room_id=room_id)
            )
            await session.commit()

        _ = await interaction.response.send_message(f'Bound channel to "{room_name}".')
    except Exception:
        _ = await interaction.response.send_message(
            f"Could not bind channel to room {room_id}."
        )
        print(traceback.format_exc())


@bot.slash_command(force_global=True)
async def unbind(interaction: nextcord.Interaction[nextcord.Client]):
    """Unbinds a Discord channel from a Content API channel.

    Parameters
    ----------
    interaction : nextcord.Interaction[nextcord.Client]
        The interaction that triggered this command.
    """
    channel_id = interaction.channel_id
    if channel_id is None:
        _ = await interaction.response.send_message(
            "This command must be run in a Discord channel."
        )
        return
    channel_id = str(channel_id)

    async with async_session.begin() as session:
        channel_pair = await session.get(ChannelPair, channel_id)
        if channel_pair is None:
            _ = await interaction.response.send_message(
                f"No channel pair found for channel {channel_id}."
            )
            return
        await session.delete(channel_pair)
        await session.commit()

    _ = await interaction.response.send_message(f"Unbound channel {channel_id}.")


async def add_attachments(content: str, message: nextcord.Message) -> str:
    ACCEPTED_MIME_TYPES = (
        "image/bmp",
        "image/gif",
        "image/jpeg",
        "image/png",
        "image/tiff",
        "image/webp",
        "image/x-portable-bitmap",
        "image/tga",
    )
    for attachment in message.attachments:
        # 25 MB
        if (
            attachment.size > 25000000
            or attachment.content_type not in ACCEPTED_MIME_TYPES
        ):
            content += f"\n!{attachment.url}"
        else:
            hash = await DiscordAttachment.get_attachment(attachment)
            content += f"\n!{content_api.file_route(hash)}"

    return content


def get_user_name(user: nextcord.Member | nextcord.User) -> str:
    if isinstance(user, nextcord.Member) and user.nick is not None:
        return str(user.nick)
    if user.global_name is not None:
        return user.global_name
    return user.name


@bot.event
async def on_message(message: nextcord.Message):
    if message.author.bot:
        return

    async with async_session.begin() as session:
        channel_pair = await session.get(ChannelPair, message.channel.id)
        if channel_pair is None:
            return

        avatar = await AvatarStore.fetch_avatar_from_user(message.author, content_api)
        # awful workaround for Column[int] not being ConvertibleToInt
        # I'll find a better solution later
        room_id = int(str(channel_pair.content_api_room_id))

        message.attachments
        content = await markup_service.discord_to_contentapi(message.content)
        content = await add_attachments(content, message)

        id = await content_api.write_message(
            room_id,
            content,
            get_user_name(message.author),
            avatar,
            "12y",
        )

        contentapi_message = ContentApiMessageStore(
            discord_message_id=message.id,
            content_api_message_id=id,
            content_api_room_id=room_id,
        )
        session.add(contentapi_message)
        await session.commit()


@bot.event
async def on_message_delete(message: nextcord.Message):
    async with async_session.begin() as session:
        contentapi_message = await session.get(ContentApiMessageStore, message.id)
        if contentapi_message is None:
            return

        await content_api.delete_message(
            int(str(contentapi_message.content_api_message_id)),
        )

        await session.delete(contentapi_message)
        await session.commit()


@bot.event
async def on_message_edit(_: nextcord.Message, after: nextcord.Message):
    async with async_session.begin() as session:
        contentapi_message = await session.get(ContentApiMessageStore, after.id)
        if contentapi_message is None:
            return

        avatar = await AvatarStore.fetch_avatar_from_user(after.author, content_api)
        content = await markup_service.discord_to_contentapi(after.content)
        content = await add_attachments(content, after)

        __ = await content_api.edit_message(
            int(str(contentapi_message.content_api_message_id)),
            int(str(contentapi_message.content_api_room_id)),
            content,
            get_user_name(after.author),
            avatar,
            "12y",
        )


@bot.event
async def on_ready():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    print("Ready!")


async def get_webhook_for_channel(
    channel_id: int, aiohttp_session: aiohttp.ClientSession
) -> nextcord.Webhook:
    async with async_session.begin() as session:
        webhook_channel_pair = await session.get(WebhookChannelPair, channel_id)
        if webhook_channel_pair is None:
            channel = await bot.fetch_channel(channel_id)
            if not isinstance(channel, TextChannel):
                raise ValueError("Channel is not a text channel")
            webhook = await channel.create_webhook(
                name=f"ContentAPI Bridge Webhook for {channel.name}",
                avatar=channel.guild.icon,
            )
            session.add(
                WebhookChannelPair(
                    discord_channel_id=channel_id, webhook_url=webhook.url
                )
            )
            await session.commit()
            return webhook
        else:
            webhook = nextcord.Webhook.from_url(
                str(webhook_channel_pair.webhook_url), session=aiohttp_session
            )
            return webhook


@content_api.on_message_created
async def on_message_created(event: contentapi.MessageEvent):
    print(event)
    if event.user is None:
        return

    if event.user.id == content_api.uid:
        return

    async with async_session.begin() as session:
        channels_pairs = await session.scalars(
            sqla.select(ChannelPair).where(
                ChannelPair.content_api_room_id == event.content_id
            )
        )

        for channel_pair in channels_pairs.all():
            async with aiohttp.ClientSession() as aiohttp_session:
                webhook = await get_webhook_for_channel(
                    int(str(channel_pair.discord_channel_id)), aiohttp_session
                )
                content = await markup_service.contentapi_to_discord(
                    event.message.text, "12y"
                )
                webhook_message = await webhook.send(
                    content=content,
                    username=event.user.name,
                    avatar_url=content_api.get_avatar(event.user),
                    wait=True,
                )
                webhook_message_data = WebhookMessageStore(
                    discord_message_id=webhook_message.id,
                    webhook_id=webhook.id,
                    webhook_message_channel_id=channel_pair.discord_channel_id,
                    contentapi_message_id=event.message.id,
                )
                session.add(webhook_message_data)
                await session.commit()


@content_api.on_message_updated
async def on_message_updated(event: contentapi.MessageEvent):
    print(event)
    if event.user is None:
        return

    if event.user.id == content_api.uid:
        return

    async with async_session.begin() as session:
        webhook_messages = await session.scalars(
            sqla.select(WebhookMessageStore).where(
                WebhookMessageStore.contentapi_message_id == event.message.id
            )
        )

        for webhook_message in webhook_messages.all():
            webhook = await bot.fetch_webhook(int(str(webhook_message.webhook_id)))
            content = await markup_service.contentapi_to_discord(
                event.message.text, "12y"
            )
            _ = await webhook.edit_message(
                int(str(webhook_message.discord_message_id)), content=content
            )


@content_api.on_message_deleted
async def on_message_deleted(event: contentapi.MessageEvent):
    print(event)
    async with async_session.begin() as session:
        webhook_messages = await session.scalars(
            sqla.select(WebhookMessageStore).where(
                WebhookMessageStore.contentapi_message_id == event.message.id
            )
        )

        for webhook_message in webhook_messages.all():
            webhook = await bot.fetch_webhook(int(str(webhook_message.webhook_id)))
            _ = await webhook.delete_message(
                int(str(webhook_message.discord_message_id))
            )

        await session.delete(webhook_messages)
        await session.commit()


def main():
    async def start():
        _ = await asyncio.gather(
            content_api.socket(),
            bot.start(os.environ["DISCORD_TOKEN"]),
        )

    asyncio.run(start())


if __name__ == "__main__":
    main()
