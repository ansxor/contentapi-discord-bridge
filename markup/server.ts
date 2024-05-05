import Markup_Parse_12y2 from "markup2/parse";
import Markup_Legacy from "markup2/legacy";
import Markup_Langs from "markup2/langs";

import { mdto12y } from "./mdto12y";
import markuprenderToMd from "./render";

const parser = new Markup_Parse_12y2();
const langs = new Markup_Langs([parser, new Markup_Legacy()]);

const server = Bun.serve({
  async fetch(req) {
    const url = new URL(req.url);
    if (url.pathname === "/discord2contentapi") {
      // discord markdown to 12y
      const body = await req.text();
      return new Response(mdto12y(body));
    } else if (url.pathname === "/contentapi2discord") {
      // markup to discord markdown
      const { searchParams } = url;
      const lang = searchParams.get("lang") ?? "12y2";
      const body = await req.text();
      const tree = langs.parse(body, lang, {});
      return new Response(markuprenderToMd(tree));
    }

    return new Response("404");
  },
});

console.log(`Running markup service on port ${server.port}!`);
