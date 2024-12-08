import aiohttp


class MarkupService:
    def __init__(self, domain: str):
        self.domain: str = domain

    async def discord_to_contentapi(self, markup: str):
        async with aiohttp.ClientSession(
            headers={"Content-Type": "text/plain"}
        ) as session:
            async with session.post(
                f"http://{self.domain}/discord2contentapi",
                data=markup,
            ) as resp:
                return await resp.text()

    async def contentapi_to_discord(self, markup: str, language: str):
        async with aiohttp.ClientSession(
            headers={"Content-Type": "text/plain"}
        ) as session:
            async with session.post(
                f"http://{self.domain}/contentapi2discord?lang={language}",
                data=markup,
            ) as resp:
                return await resp.text()
