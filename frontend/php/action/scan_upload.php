<?php
// SPDX-License-Identifier: MIT
// scan_upload.php — action handler + panel for "upload a scan-result JSON
// file". The handler POSTs the file contents to SCAN_IMPORT_FILE_ENDPOINT
// (/scan/import-file). On success the operator gets a summary message.
//
// Wire-up: include from importscan.php after the dispatcher; render the
// panel with scan_upload_panel_html() in the same file.

require_once __DIR__ . '/import_common.php';

// --- handler ----------------------------------------------------------------

function do_upload(&$importMessage) {
    if (!isset($_FILES['scanfile']) || $_FILES['scanfile']['error'] !== UPLOAD_ERR_OK) {
        $importMessage = 'Please choose a valid scan-result JSON file.';
        return;
    }
    $contents = file_get_contents($_FILES['scanfile']['tmp_name']);
    if ($contents === false || trim($contents) === '') {
        $importMessage = 'Uploaded file is empty.';
        return;
    }
    if (json_decode($contents) === null && json_last_error() !== JSON_ERROR_NONE) {
        $importMessage = 'Uploaded file is not valid JSON.';
        return;
    }
    list($code, $body, $err) = api_post_json(SCAN_IMPORT_FILE_ENDPOINT, $contents);
    $importMessage = ($code === 200)
        ? (json_decode($body, true)['message'] ?? 'Import completed.')
        : 'Import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
}

// --- panel -------------------------------------------------------------------

function scan_upload_panel_html() {
    return ''
        . '<div class="box" style="margin-top:24px;">'
        .     '<h3>Upload scan-result JSON</h3>'
        .     '<p style="color:#666; font-size:0.9em;">Upload a JSON file produced by <code>nsl-graph scan</code> (same format as <code>scan import &lt;file&gt;</code>).</p>'
        .     '<form method="post" action="import.php" enctype="multipart/form-data">'
        .         '<input type="file" name="scanfile" accept=".json,application/json" required>'
        .         '<button type="submit" name="do_upload" value="1">Upload &amp; import</button>'
        .     '</form>'
        . '</div>';
}