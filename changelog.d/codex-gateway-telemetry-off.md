- Codex gateway launches render three more owned `-c` pairs after
  `features.apps=false`: `features.plugins=false`, `analytics.enabled=false`
  and `check_for_update_on_startup=false`, in that order. With its defaults
  the Codex client refreshes its remote plugin catalog from the vendor at
  every start (83 MB in one measured launch), keeps a metrics connection to
  the vendor that posts once a minute, and looks up a newer release; all of
  it went beside the gateway. A gateway launch now runs without the plugins
  installed in the Codex home. The renderer's own check expects twelve owned
  arguments and refuses an override of `features`, `analytics`, `otel` or
  `check_for_update_on_startup` after them, and every native spelling of such
  a selection stays refused at admission. Launches without a gateway binding
  are unchanged. A consumer that compared the first six final arguments must
  expect twelve.
