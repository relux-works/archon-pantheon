- Codex gateway launches render three more owned `-c` pairs after
  `features.apps=false`: `features.plugins=false`, `analytics.enabled=false`
  and `check_for_update_on_startup=false`, in that order, after the provider
  table and the provider selection. The final-argument verifier requires
  those settings in the rendered argv. The tests do not launch Codex and do
  not establish that a client release applies the settings or suppresses the
  associated traffic. The update-check behavior has not been witnessed. The
  `gh` status-line child remains outside these switches. The gateway argv
  includes `features.plugins=false` for the remote-catalog suppression goal;
  only that argv value is verified, not the behavior of installed Codex
  releases or of future local plugin setups. One gateway launch that left the
  three new settings at the client's defaults, measured before this change,
  received 83 MB on one direct connection and posted on another once a
  minute; that is an observation of a default-configuration launch, not of
  these settings. The renderer's own check expects twelve owned arguments and
  refuses an override of `features`, `analytics`, `otel` or
  `check_for_update_on_startup` after them, in the `-c=key=value` spelling
  too, and every native spelling of such a selection stays refused at
  admission. Launches without a gateway binding are unchanged. A consumer
  that compared the first six final arguments must expect twelve.
