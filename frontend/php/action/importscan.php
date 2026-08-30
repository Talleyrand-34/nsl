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
require_once __DIR__ . '/scan.php';
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

// Saved profiles for the dropdowns — fetched on every request so the
// Profile picker in the Live scan panel is populated on a fresh POST.
$profiles = [];
$raw = @file_get_contents(SCAN_PROFILES_ENDPOINT);
$profiles = json_decode($raw, true) ?: [];
if (!is_array($profiles)) {
    $profiles = [];
}


$genericProfiles = array_values(array_filter(
    $profiles,
    fn($p) => ($p['kind'] ?? '') === 'generic',
));


// Available OS / firmware types for the profile-level OS dropdown, the
// per-device OS override dropdown and the Live scan SSH OS selector. The
// catalogue is small (~6 entries) so fetching it on every request is cheap.
$osTypes = [];
$raw = @file_get_contents(OSTYPES_ENDPOINT);
$osTypes = json_decode($raw, true) ?: [];
if (!is_array($osTypes)) {
    $osTypes = [];
}

// Profile-loader overwrites $pf when the operator clicks Load.
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_load_profile'])) {
    do_load_profile($profileMessage, $profiles, $pf);
}

// Clean-scan: drop discovered + imported and redirect.
if (isset($_GET['clear'])) {
    unset(
        $_SESSION['scan_discovered'],
        $_SESSION['scan_imported'],
        $_SESSION['connections_result'],
        $_SESSION['device_scan_id'],
        $_SESSION['connection_scan_id'],
        $_SESSION['scan_followup'],
        $_SESSION['scan_log']
    );
    $discovered = [];
    $importedIPs = [];
    $result = null;
}

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
        // Phase 1 of SCAN-REFACTOR: the unified Live scan panel chains the
        // connection scan internally. The legacy standalone button is
        // gone; this branch is a no-op kept for back-compat with any
        // unsaved form state.
    }
}

if ($_SERVER['REQUEST_METHOD'] === 'GET' && isset($_GET['scan_id'])) {
    $scanRunId = do_scan_completed($discovered, $scanMessage, $plan, $importedIPs, $importMessage, $result);
    // Connection follow-up (when also_connections was set) is handled inside
    // do_scan_completed; it returns whichever scan is still running so the
    // status panel keeps watching it.
}

// ---------------------------------------------------------------------------
// Render: two-column grid.
//   left  = scan options (collapsed accordion: only one card open)
//   right = scan status + results + upload
// ---------------------------------------------------------------------------
?>
<h2>Import devices</h2>


<?php if ($scanMessage): ?>
  <p class="flash flash-error"><?= htmlspecialchars($scanMessage) ?></p>
<?php endif; ?>

<div class="grid import-grid">
  <!-- Left column: scan options + saved profiles + create profile. The accordion
       JS handles one-open-at-a-time within this column. -->
  <div class="actions import-actions">
    <section class="accordion-card" data-accordion="import-actions">
      <h3 class="accordion-header"><button type="button" class="accordion-toggle" aria-expanded="true">Live scan</button></h3>
      <div class="accordion-body">
        <?= scan_live_panel_html($profiles, $scanMessage, $osTypes) ?>
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
        <?= scan_profiles_create_panel_html($profiles, $osTypes) ?>
      </div>
    </section>

  </div><!-- /.actions.import-actions -->

  <!-- Right column: scan status + results + upload. -->
  <div class="results import-results">

    <?php if ($importMessage): ?>
      <p class="flash flash-success"><?= htmlspecialchars($importMessage) ?></p>
    <?php endif; ?>

    <?= scan_status_panel_html($scanRunId, json_encode($autoReload)) ?>
    <?php if ($profileMessage): ?>
      <p class="flash flash-info"><?= htmlspecialchars($profileMessage) ?></p>
    <?php endif; ?>

    <?php if ($scanMessage): ?>
      <p class="flash flash-error"><?= htmlspecialchars($scanMessage) ?></p>
    <?php endif; ?>

    <?= scan_results_panel_html($discovered, $importedIPs, $plan, $result, $scanMessage) ?>

    <!-- Inline upload form lives here per the user's request: it feeds
         directly into the same results column. -->
    <section class="results-upload">
      <h3>Upload scan-result JSON</h3>
      <p class="hint">Upload a JSON file produced by <code>nsl-graph scan</code> (same format as <code>scan import &lt;file&gt;</code>).</p>
      <form method="post" action="import.php" enctype="multipart/form-data">
        <input type="file" name="scanfile" accept=".json,application/json" required>
        <button type="submit" name="do_upload" value="1">Upload &amp; import</button>
      </form>

    <?= scan_activity_panel_html() ?>
  </div>
</div>

<script src="import-accordion.js"></script>