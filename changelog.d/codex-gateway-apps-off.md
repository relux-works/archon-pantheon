- Codex gateway launches render `-c features.apps=false` as a third owned
  pair after the provider table and the provider selection. The Codex client
  adds a ChatGPT-hosted MCP server (`codex_apps`) by itself when the `apps`
  feature is on, which is its default; app tool calls and the files uploaded
  for them went to the vendor beside the gateway. The renderer's own check
  now expects six owned arguments and refuses a feature selector after them,
  and every native spelling of a feature selection stays refused at
  admission. Launches without a gateway binding are unchanged. A consumer
  that compared the first four final arguments must expect six.
