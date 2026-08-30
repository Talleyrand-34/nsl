<?php
// SPDX-License-Identifier: MIT
// import_common.php — shared helpers for the unified import page (and
// its backward-compat redirect from the old connections.php URL).
//
// The dispatcher in importscan.php plus the per-concern action modules
// (scan_devices, scan_connections, scan_profiles, scan_upload)
// all share this file. Function definitions are wrapped in
// `function_exists` guards so multiple includes are no-ops.
//
// All functions are HTTP-only and stateless. Callers own the resulting
// `$scanMessage`, `$importMessage`, etc. message strings.

/**
 * api_post_json POSTs a JSON string to the API.
 * Returns [httpCode, body, curlError]. The longer default timeout suits
 * scan POSTs which can take minutes for large subnets.
 */
if (!function_exists('api_post_json')) {
    function api_post_json($url, $jsonBody, $timeoutSec = 280) {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
        curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonBody);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $timeoutSec);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($jsonBody),
        ]);
        $body = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $err  = curl_error($ch);
        curl_close($ch);
        return [$code, $body, $err];
    }
}

/**
 * api_get GETs a URL. Returns [httpCode, body].
 * Used by both pages for /scan/status polling and profile reads.
 */
if (!function_exists('api_get')) {
    function api_get($url, $timeoutSec = 15) {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $timeoutSec);
        $body = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        return [$code, $body];
    }
}

/**
 * api_method issues a request with an optional JSON body; returns [code, body].
 * Kept here for callers that need a verb other than POST (DELETE for
 * profile removal, GET for status). The old `api_method` from importscan.php
 * is replaced by this one; call sites are identical.
 */
if (!function_exists('api_method')) {
    function api_method($method, $url, $jsonBody = null) {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        $headers = [];
        if ($jsonBody !== null) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonBody);
            $headers[] = 'Content-Type: application/json';
            $headers[] = 'Content-Length: ' . strlen($jsonBody);
        }
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
        $body = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        return [$code, $body];
    }
}

/**
 * scan_status_panel_html renders the live scan-status panel that
 * nslWatchScan() drives from /scan/status polling. Returns the same
 * DOM contract both pages already use (#scan-status, #scan-state,
 * #scan-bar, #scan-bar-fill, #scan-events) so the existing JS works
 * unchanged.
 *
 * Call once per page when $scanRunId is set. Output is bare — no
 * surrounding wrapper. $title lets each caller customise the heading
 * ("Scanning…", "Discovering connections…", etc.).
 */
if (!function_exists('scan_status_panel_html')) {
    function scan_status_panel_html($scanRunId, $autoReloadJson, $title = 'Scanning…') {
        if (empty($scanRunId)) {
            return '';
        }
        return ''
            . '<div id="scan-status" class="scan-status">'
            .     '<h4><span class="spinner"></span>' . htmlspecialchars($title) . '</h4>'
            .     '<div class="scan-state" id="scan-state">starting…</div>'
            .     '<div class="scan-bar" id="scan-bar" style="display:none;"><div class="scan-bar-fill" id="scan-bar-fill"></div></div>'
            .     '<div class="scan-events" id="scan-events"></div>'
            . '</div>'
            . '<script>nslWatchScan(' . json_encode($scanRunId)
            . ', {reloadParams: ' . $autoReloadJson . '});</script>';
    }
}

/**
 * scan_log_push appends a phase record to $_SESSION['scan_log'] (newest
 * first, capped). Use this from every scan-pipeline handler so that
 * failed or empty phases show up in the activity panel instead of
 * silently disappearing.
 */
if (!function_exists('scan_log_push')) {
    function scan_log_push($phase, $runId, $state, $detail, $title = '') {
        if (!isset($_SESSION['scan_log']) || !is_array($_SESSION['scan_log'])) {
            $_SESSION['scan_log'] = [];
        }
        array_unshift($_SESSION['scan_log'], [
            'ts'     => date('Y-m-d\TH:i:sP'),
            'phase'  => (string) $phase,
            'run_id' => (string) $runId,
            'state'  => (string) $state,
            'detail' => (string) $detail,
            'title'  => (string) $title,
        ]);
        $_SESSION['scan_log'] = array_slice($_SESSION['scan_log'], 0, 20);
    }
}

/**
 * scan_log_events fetches /scan/status?scan_id= for a finished run and
 * returns the granular events. Cached in the log record on first call so
 * subsequent renders are free.
 */
if (!function_exists('scan_log_events')) {
    function scan_log_events(&$record) {
        if (!empty($record['_events'])) {
            return $record['_events'];
        }
        if (empty($record['run_id'])) {
            return [];
        }
        list($code, $body) = api_get(SCAN_STATUS_ENDPOINT . '?scan_id=' . urlencode($record['run_id']), 10);
        if ($code !== 200) {
            return [];
        }
        $status = json_decode($body, true) ?: [];
        $events = $status['events'] ?? [];
        if (!empty($events)) {
            $record['_events'] = $events;
        }
        return $events;
    }
}

/**
 * scan_activity_panel_html renders the "Scan activity & errors" card from
 * $_SESSION['scan_log']. Always emits a wrapper so the section is visible
 * even when empty (the no-activity message is the empty state).
 */
if (!function_exists('scan_activity_panel_html')) {
    function scan_activity_panel_html() {
        $log = $_SESSION['scan_log'] ?? [];
        ob_start();
        ?>
        <section class="scan-activity" style="margin-top:18px;">
          <h3>Scan activity &amp; errors</h3>
          <?php if (empty($log)): ?>
            <p style="color:#777; font-style:italic; margin:4px 0;">No scan activity yet. Run a scan from the Live scan panel.</p>
          <?php else: ?>
            <table border="1" cellpadding="4" cellspacing="0" style="font-size:0.9em;">
              <thead><tr><th>When</th><th>Phase</th><th>State</th><th>Detail</th><th>Run</th><th></th></tr></thead>
              <tbody>
                <?php foreach ($log as $i => $r):
                    $row = $r;
                    $events = scan_log_events($row);
                    $stateClass = 'sa-' . preg_replace('/[^a-z]/', '', strtolower($row['state'] ?? ''));
                    $rowId = 'sa-row-' . $i;
                ?>
                  <tr class="<?= htmlspecialchars($stateClass) ?>">
                    <td><?= htmlspecialchars($row['ts']) ?></td>
                    <td><?= htmlspecialchars($row['phase']) ?></td>
                    <td><strong><?= htmlspecialchars($row['state']) ?></strong></td>
                    <td><?= htmlspecialchars($row['detail']) ?></td>
                    <td><?= htmlspecialchars($row['run_id']) ?><?= $row['title'] ? ' &middot; ' . htmlspecialchars($row['title']) : '' ?></td>
                    <td>
                      <?php if (!empty($row['run_id'])): ?>
                        <details id="<?= $rowId ?>"><summary>events</summary>
                          <?php if (empty($events)): ?>
                            <p style="color:#777; font-style:italic; margin:4px 0;">Run not found in the backend registry (it may have been recycled or the server restarted).</p>
                          <?php else: ?>
                            <div style="max-height:240px; overflow:auto; font-family:monospace; font-size:0.8em; background:#fafafa; padding:6px; border:1px solid #eee;">
                              <?php foreach ($events as $e):
                                $level = htmlspecialchars($e['level'] ?? 'info');
                                $msg = htmlspecialchars($e['msg'] ?? '');
                                $fields = $e['fields'] ?? [];
                                $fstr = '';
                                foreach ($fields as $k => $v) {
                                    $fstr .= ' ' . htmlspecialchars((string) $k) . '=' . htmlspecialchars(is_scalar($v) ? (string) $v : json_encode($v));
                                }
                              ?>
                                <div>[<?= $level ?>] <?= $msg ?><?= $fstr ?></div>
                              <?php endforeach; ?>
                            </div>
                          <?php endif; ?>
                        </details>
                      <?php endif; ?>
                    </td>
                  </tr>
                <?php endforeach; ?>
              </tbody>
            </table>
          <?php endif; ?>
        </section>
        <?php
        return ob_get_clean();
    }
}
