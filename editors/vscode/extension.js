// The RuneScript extension: the grammar and editor settings are declared
// in package.json; this file starts the language server, `runescript lsp`,
// which gives diagnostics as you type, hover, go to definition,
// completion, and the outline.
"use strict";

const vscode = require("vscode");
const { LanguageClient, TransportKind } = require("vscode-languageclient/node");

let client;

async function startServer() {
  const config = vscode.workspace.getConfiguration("runescript");
  if (!config.get("server.enabled", true)) {
    return;
  }
  const command = config.get("server.path") || "runescript";
  const serverOptions = {
    command,
    args: ["lsp"],
    transport: TransportKind.stdio,
  };
  const clientOptions = {
    documentSelector: [
      { scheme: "file", language: "runescript" },
      { scheme: "untitled", language: "runescript" },
    ],
  };
  client = new LanguageClient("runescript", "RuneScript", serverOptions, clientOptions);
  try {
    await client.start();
  } catch (err) {
    client = undefined;
    const choice = await vscode.window.showErrorMessage(
      `RuneScript: could not start the language server with "${command} lsp" (${err.message}). ` +
        "Install runescript, or set runescript.server.path to the command.",
      "Open Settings"
    );
    if (choice === "Open Settings") {
      vscode.commands.executeCommand("workbench.action.openSettings", "runescript.server");
    }
  }
}

async function stopServer() {
  if (client) {
    const c = client;
    client = undefined;
    await c.stop();
  }
}

async function activate(context) {
  context.subscriptions.push(
    vscode.commands.registerCommand("runescript.restartServer", async () => {
      await stopServer();
      await startServer();
    }),
    vscode.workspace.onDidChangeConfiguration(async (e) => {
      if (e.affectsConfiguration("runescript.server")) {
        await stopServer();
        await startServer();
      }
    })
  );
  await startServer();
}

function deactivate() {
  return stopServer();
}

module.exports = { activate, deactivate };
