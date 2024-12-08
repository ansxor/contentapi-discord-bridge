import unittest

from nextcord import User

from contentapi_discord_bridge import contentapi


class TestContentApi(unittest.TestCase):
    def test_get_avatar(self):
        user = contentapi.User(id=12, name="answer", avatar="jxoqo")
        capi = contentapi.ContentApi("example", "token")
        self.assertEqual(
            capi.get_avatar(user),
            "https://example/api/File/raw/jxoqo?size=100&crop=true",
        )
        self.assertEqual(
            capi.get_avatar(user, 200),
            "https://example/api/File/raw/jxoqo?size=200&crop=true",
        )

    def test_parse_single_message_event(self):
        data = '{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"a":"jxoqo","n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}'
        events = contentapi.ContentApi.parse_message_events(data)

        if len(events) != 1:
            self.fail("ParseMessageEvents failed")

        event = events[0]

        self.assertEqual(event.state, contentapi.MessageEventType.CREATED)
        self.assertEqual(event.content_id, 6661)
        if event.user is None:
            self.fail("User is None")
        else:
            self.assertEqual(event.user.id, 12)
            self.assertEqual(event.user.name, "answer")
            self.assertEqual(event.user.avatar, "jxoqo")
        self.assertEqual(event.message.id, 1239520)
        self.assertEqual(event.message.text, "meow")
        self.assertEqual(event.message.markup, "12y2")

    def test_parse_single_edited_message_event(self):
        data = """{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5292,"events":[{"id":5292,"date":"2023-10-14T13:44:18.97Z","userId":12,"action":4,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meowasd","editDate":"2023-10-14T13:44:18.972Z","editUserId":12,"edited":1,"deleted":0,"module":null,"receiveUserId":0,"values":{"a":"jxoqo","n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}"""
        events = contentapi.ContentApi.parse_message_events(data)

        if len(events) != 1:
            self.fail("ParseMessageEvents failed")

        event = events[0]

        self.assertEqual(event.state, contentapi.MessageEventType.UPDATED)
        self.assertEqual(event.content_id, 6661)
        self.assertIsNotNone(event.user)
        if event.user is None:
            self.fail("User is None")
        else:
            self.assertEqual(event.user.id, 12)
            self.assertEqual(event.user.name, "answer")
            self.assertEqual(event.user.avatar, "jxoqo")
        self.assertEqual(event.message.id, 1239520)
        self.assertEqual(event.message.text, "meowasd")
        self.assertEqual(event.message.markup, "12y2")

    def test_parse_single_deleted_message_event(self):
        data = """{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5293,"events":[{"id":5293,"date":"2023-10-14T13:45:23.06Z","userId":12,"action":8,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":0,"createDate":"2023-10-14T13:35:40.359Z","text":"deleted_comment","editDate":"2023-10-14T13:44:18.972Z","editUserId":0,"edited":0,"deleted":1,"module":null,"receiveUserId":0,"values":{},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}"""
        events = contentapi.ContentApi.parse_message_events(data)

        if len(events) != 1:
            self.fail("ParseMessageEvents failed")

        event = events[0]
        self.assertEqual(event.state, contentapi.MessageEventType.DELETED)
        self.assertEqual(event.message.id, 1239520)
        self.assertEqual(event.message.text, "deleted_comment")

    def test_parse_with_message_avatar(self):
        data = """{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"a":"meows","n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}"""
        events = contentapi.ContentApi.parse_message_events(data)

        if len(events) != 1:
            self.fail("ParseMessageEvents failed")

        event = events[0]

        if event.user is None:
            self.fail("User is None")
        self.assertEqual(event.user.avatar, "meows")

    def test_parse_with_user_avatar(self):
        data = """{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"n":"do you feel my worth?","m":"12y2"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}"""
        events = contentapi.ContentApi.parse_message_events(data)

        if len(events) != 1:
            self.fail("ParseMessageEvents failed")

        event = events[0]

        if event.user is None:
            self.fail("User is None")
        self.assertEqual(event.user.avatar, "jxoqo")

    def test_parse_with_no_markup(self):
        data = """{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{"message_event":{"message":[{"id":1239520,"contentId":6661,"createUserId":12,"createDate":"2023-10-14T13:35:40.359Z","text":"meow","editDate":null,"editUserId":null,"edited":0,"deleted":0,"module":null,"receiveUserId":0,"values":{"n":"do you feel my worth?"},"uidsInText":[],"engagement":{}}],"content":[{"id":6661,"name":"XXXX","parentId":645,"lastRevisionId":45197,"createDate":"2021-03-13T04:02:47.000Z","createUserId":12,"deleted":0,"contentType":1,"literalType":"chat","hash":"6661","values":{"system":"switch","markupLang":"plaintext"},"permissions":{"12":"CRUD","5410":"CRUD"}}],"user":[{"id":12,"username":"answer","avatar":"jxoqo","special":"","type":1,"createDate":"2020-04-23T04:04:24.000Z","createUserId":0,"super":0,"registered":1,"deleted":0,"groups":[31851,31853],"usersInGroup":[]}]}}},"error":null}"""
        events = contentapi.ContentApi.parse_message_events(data)

        if len(events) != 1:
            self.fail("ParseMessageEvents failed")

        event = events[0]

        self.assertEqual(event.message.markup, "plaintext")

    def test_live_event_without_message_events(self):
        data = """{"id":"","type":"live","requestUserId":12,"data":{"optimized":true,"lastId":5291,"events":[{"id":5291,"date":"2023-10-14T13:35:40.37Z","userId":12,"action":1,"type":"message_event","refId":1239520,"contentId":0}],"objects":{}},"error":null}"""
        events = contentapi.ContentApi.parse_message_events(data)

        if len(events) != 0:
            self.fail("ParseMessageEvents failed")


if __name__ == "__main__":
    _ = unittest.main()
