export const escapeXml = (text: string) =>
  text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");

export const escape12y = (text: string) =>
  text.replace(/[\\\/\{\*>_~`]/g, "\\$&");

export const escapeMd = (text: string) =>
  String(text).replace(/[\\*`_~]/g, "\\$&");
