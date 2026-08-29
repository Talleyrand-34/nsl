<?php
// SPDX-License-Identifier: MIT
// connections.php — backward-compat redirect.
//
// After phase 3 the connection scan lives inside import.php's unified
// page. This file used to render its own panel; now it's a 302 to
// import.php so any existing bookmark / link keeps working without a
// stale page.
//
// Drop this file once all callers are confirmed migrated (typically a
// sprint after the unified page ships).
header('Location: import.php', true, 307);
exit;