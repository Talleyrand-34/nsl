<?php
// SPDX-License-Identifier: MIT
// import_common.php — shared helpers for the unified import page (and
// its backward-compat redirect from the old connections.php URL).
//
// The dispatcher in importscan.php plus the per-concern action modules
// (scan_devices, scan_connections, scan_profiles, scan_upload, scan_bulk)
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
            .   ', {reloadParams: ' . $autoReloadJson . '});</script>';
    }
}