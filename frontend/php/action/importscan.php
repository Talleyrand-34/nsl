<?php
// SPDX-License-Identifier: MIT
// importscan.php — dispatcher + two-column layout for the unified import
// page.
//
// After phase 3 the page renders as a two-column grid matching main.php:
//   .grid > .actions  — left column: scan options (Live scan, Bulk,
//                      Saved profiles, Create profile, Live conn scan).
//   .grid > .results  — right column: live status panel, discovered-
//                      devices table, plan review, inline upload form,
//                      connection scan results.
//
// Each per-concern handler lives in its own action file; the dispatcher
// only routes by $_POST[do_X].

require_once __DIR__ . '/import_common.php';
require_once __DIR__ . '/scan_profiles.php';
require_once __DIR__ . '/scan_devices.php';
require_once __DIR__ . '/scan_connections.php';
require_once __DIR__ . '/scan_upload.php';


// ---------------------------------------------------------------------------
// State vars. Handlers mutate these in place via PHP's pass-by-ref.
// ---------------------------------------------------------------------------

$scanMessage    = '';
$importMessage  = '';
$profileMessage = '';
$plan           = null;
$scanRunId      = '';
$autoReload     = '';

// Discovered devices + which IPs were already imported persist in the
// session across the analyze/import round-trips.
$discovered  = $_SESSION['scan_discovered'] ?? [];
$importedIPs = $_SESSION['scan_imported']   ?? [];
$result     = $_SESSION['connections_result'] ?? null;

// SNMP form prefill (from a loaded profile, else POST defaults).
$pf = [
    'target'       => '',
    'community'    => 'public',
    'snmp_version' => '2c',
    'snmp_port'    => '161',
];

// Saved profiles for the dropdowns — loaded lazily on form-related posts.
$profiles = [];
if ($_SERVER['REQUEST_METHOD'] === 'GET'
    || isset($_POST['do_create_profile'])
    || isset($_POST['do_delete_profile'])
    || isset($_POST['do_load_profile'])) {
    $raw = @file_get_contents(SCAN_PROFILES_ENDPOINT);
    $profiles = json_decode($raw, true) ?: [];
}
if (!is_array($profiles)) {
    $profiles = [];
}
$genericProfiles = array_values(array_filter(
    $profiles,
    fn($p) => ($p['kind'] ?? '') === 'generic',
));

// Profile-loader overwrites $pf when the operator clicks Load.
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_load_profile'])) {
    do_load_profile($profileMessage, $profiles, $pf);
}

// Clean-scan: drop discovered + imported and redirect.
if (isset($_GET['clear'])) {
    unset($_SESSION['scan_discovered'], $_SESSION['scan_imported'], $_SESSION['connections_result']);
    $discovered = [];
    $importedIPs = [];
    $result = null;
}

// ---------------------------------------------------------------------------
// Dispatch.
// ---------------------------------------------------------------------------

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['do_create_profile'])) {
        do_create_profile($profileMessage, $profiles);
    }
    if (isset($_POST['do_delete_profile'])) {
        do_delete_profile($profileMessage, $profiles);
    }
    if (isset($_POST['do_scan'])) {
        do_scan($scanMessage, $scanRunId, $autoReload);
    }
    if (isset($_POST['do_analyze'])) {
        do_analyze($plan, $scanMessage);
    }
    if (isset($_POST['do_execute'])) {
        do_execute($plan, $importMessage, $importedIPs);
    }
    if (isset($_POST['do_upload'])) {
        do_upload($importMessage);
    }

    if (isset($_POST['do_scan_connections'])) {
        do_scan_connections($scanMessage, $scanRunId);
    }
}

// GET ?scan_id= — pull scan status. do_scan_completed returns true ONLY
// when the scan is still running; otherwise we clear $scanRunId so the
// page renders the static results panel without re-arming nslWatchScan.
if ($_SERVER['REQUEST_METHOD'] === 'GET' && isset($_GET['scan_id'])) {
    $deviceRunning = do_scan_completed($discovered, $scanMessage, $plan, $importedIPs, $importMessage);
    $scanRunId = $deviceRunning ? $_GET['scan_id'] : '';
    do_scan_connections_completed($scanMessage, $scanRunId, $result);
}

// ---------------------------------------------------------------------------
// Render: two-column grid.
//   left  = scan options (collapsed accordion: only one card open)
//   right = scan status + results + upload
// ---------------------------------------------------------------------------
?>
<h2>Import devices</h2>

<div class="grid import-grid">
  <!-- Left column: scan options. The accordion JS handles one-open-at-a-time. -->
  <div class="actions import-actions">

    <section class="accordion-card" data-accordion="import-actions">
      <h3 class="accordion-header"><button type="button" class="accordion-toggle" aria-expanded="true">Live device scan</button></h3>
      <div class="accordion-body">
        <?= scan_devices_live_panel_html($pf, $profiles, $scanMessage) ?>
      </div>
    </section>

    <section class="accordion-card" data-accordion="import-actions">
      <h3 class="accordion-header"><button type="button" class="accordion-toggle" aria-expanded="false">Live connection scan</button></h3>
      <div class="accordion-body">
        <?= scan_connections_live_panel_html($genericProfiles, $scanMessage) ?>
      </div>
    </section>



    <section class="accordion-card" data-accordion="import-actions">
      <h3 class="accordion-header"><button type="button" class="accordion-toggle" aria-expanded="false">Saved scan profiles (<?= count($profiles) ?>)</button></h3>
      <div class="accordion-body">
        <?= scan_profiles_saved_panel_html($profiles) ?>
      </div>
    </section>

    <section class="accordion-card" data-accordion="import-actions">
      <h3 class="accordion-header"><button type="button" class="accordion-toggle" aria-expanded="false">+ Create a new profile</button></h3>
      <div class="accordion-body">
        <?= scan_profiles_create_panel_html($profiles) ?>
      </div>
    </section>

  </div>

  <!-- Right column: scan status + results + upload. -->
  <div class="results import-results">

    <?php if ($importMessage): ?>
      <p class="flash flash-success"><?= htmlspecialchars($importMessage) ?></p>
    <?php endif; ?>

    <?= scan_status_panel_html($scanRunId, json_encode($autoReload)) ?>

    <?php if ($profileMessage): ?>
      <p class="flash flash-info"><?= htmlspecialchars($profileMessage) ?></p>
    <?php endif; ?>

    <?= scan_devices_discovered_panel_html($discovered, $importedIPs) ?>

    <?= scan_devices_plan_panel_html($plan) ?>

    <?= scan_connections_results_panel_html($result, $scanMessage) ?>

    <!-- Inline upload form lives here per the user's request: it feeds
         directly into the same results column. -->
    <section class="results-upload">
      <h3>Upload scan-result JSON</h3>
      <p class="hint">Upload a JSON file produced by <code>nsl-graph scan</code> (same format as <code>scan import &lt;file&gt;</code>).</p>
      <form method="post" action="import.php" enctype="multipart/form-data">
        <input type="file" name="scanfile" accept=".json,application/json" required>
        <button type="submit" name="do_upload" value="1">Upload &amp; import</button>
      </form>
    </section>

  </div>
</div>

<script src="import-accordion.js"></script>
<script src="pending-devices.js"></script>