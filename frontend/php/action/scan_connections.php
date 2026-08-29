<?php
// SPDX-License-Identifier: MIT
// scan_connections.php — connection-scan handlers + panels.
//
// Handlers:
//   do_scan_connections        — async POST to /scan/connections, returns scan_id.
//   do_scan_connections_completed — pull /scan/status for ?scan_id= and
//                                  populate $result / $scanRunId on completion.
//   do_import_connections      — POST selected edges to
//                                  /scan/connections/import. STUB for phase 3
//                                  (port-editor machinery lives in the legacy
//                                  connections.php for now).
//
// Panels:
//   scan_connections_live_panel_html      — the scan form (mode + SSH +
//                                            collector + timeout).
//   scan_connections_results_panel_html   — discovered edges table +
//                                            inline topology diagram.
//
// State carried by reference:
//   $scanMessage  — human message
//   $scanRunId    — current run id (set on submit, used to render the live
//                   panel + clear it after completion so the page doesn't
//                   re-arm the watcher)
//   $result       — full connection-scan result ($_SESSION['connections_result'])
//   $importedIPs  — list of newly-imported edge IDs (currently unused;
//                   do_import_connections is a stub)

require_once __DIR__ . '/import_common.php';

// --- handlers ---------------------------------------------------------------

function scan_connections_form_state() {
    return [
        'mode'            => $_POST['mode'] ?? 'from-db',
        'subnet'          => $_POST['subnet'] ?? '',
        'community'       => $_POST['community'] ?? 'public',
        'collector'       => $_POST['collector'] ?? '',
        'timeout'         => $_POST['timeout'] ?? '10',
        'ssh_user'        => $_POST['ssh_user'] ?? '',
        'ssh_key'         => $_POST['ssh_key'] ?? '',
        'ssh_password'    => $_POST['ssh_password'] ?? '',
        'generic_profile' => $_POST['generic_profile'] ?? '',
    ];
}

function do_scan_connections(&$scanMessage, &$scanRunId) {
    // Uploaded OpenSSH config + key files are read into memory and sent in
    // the request body — never written to disk on this host.
    $sshConfig = '';
    if (isset($_FILES['ssh_config']) && $_FILES['ssh_config']['error'] === UPLOAD_ERR_OK) {
        $sshConfig = (string) file_get_contents($_FILES['ssh_config']['tmp_name']);
    }
    $sshKeys = [];
    $keyCollision = '';
    if (isset($_FILES['ssh_keys']) && is_array($_FILES['ssh_keys']['name'])) {
        foreach ($_FILES['ssh_keys']['name'] as $idx => $fname) {
            if (($_FILES['ssh_keys']['error'][$idx] ?? UPLOAD_ERR_NO_FILE) !== UPLOAD_ERR_OK) {
                continue;
            }
            $base = basename($fname); // path ignored; match by basename (as ssh-config IdentityFile)
            if (isset($sshKeys[$base])) {
                $keyCollision = $base;
                break;
            }
            $sshKeys[$base] = (string) file_get_contents($_FILES['ssh_keys']['tmp_name'][$idx]);
        }
    }

    if ($keyCollision !== '') {
        $scanMessage = 'Two uploaded key files share the basename "' . htmlspecialchars($keyCollision)
            . '". ssh-config matches keys by basename, so names must be unique. Rename one and retry.';
        return;
    }

    $f = scan_connections_form_state();
    $opts = [
        'from_db'         => $f['mode'] === 'from-db',
        'profiles'        => $f['mode'] === 'profiles',
        'subnet'          => $f['mode'] === 'subnet' ? trim($f['subnet']) : '',
        'community'       => $f['community'],
        'collector'       => $f['collector'],
        'timeout_sec'     => intval($f['timeout']),
        'ssh_user'        => $f['ssh_user'],
        'ssh_key_file'    => $f['ssh_key'],
        'ssh_password'    => $f['ssh_password'],
        'generic_profile' => $f['generic_profile'],
        'ssh_config'      => $sshConfig,
        // Cast to object so an empty set encodes as {} (a JSON object), not []
        // — the API expects a map for ssh_keys.
        'ssh_keys'        => (object) $sshKeys,
    ];
    // The scan runs async now: this returns a scan_id immediately and the
    // live panel polls /scan/status, reloading with ?scan_id= when complete.
    list($code, $body, $err) = api_post_json(SCAN_CONNECTIONS_ENDPOINT, json_encode($opts), 30);
    $j = json_decode($body, true);
    if (($code === 202 || $code === 200) && !empty($j['scan_id'])) {
        $scanRunId = $j['scan_id'];
    } else {
        $scanMessage = 'Could not start scan (HTTP ' . intval($code) . '): ' . htmlspecialchars($body . ' ' . $err);
    }
}

function do_scan_connections_completed(&$scanMessage, &$scanRunId, &$result) {
    // GET ?scan_id= → /scan/status. When complete, populate $result from
    // the snapshot's `edges` / `hosts` / `intermediaries` so the page
    // can render the discovered topology.
    if (!isset($_GET['scan_id'])) {
        return false;
    }
    list($code, $body) = api_method('GET', SCAN_STATUS_ENDPOINT . '?scan_id=' . urlencode($_GET['scan_id']));
    $st = json_decode($body, true);
    $state = is_array($st) ? ($st['state'] ?? '') : '';
    if ($code === 200 && $state === 'completed') {
        $result = is_array($st) ? ($st['result'] ?? null) : null;
        if ($result) {
            $_SESSION['connections_result'] = $result;
        }
        $scanMessage = is_array($result) && !empty($result['edges'])
            ? count($result['edges']) . ' edge(s) discovered.'
            : 'Scan completed; no edges to import.';
        return false;
    }
    if ($code === 200 && $state === 'failed') {
        $scanMessage = 'Connection scan failed: ' . htmlspecialchars($st['error'] ?? 'unknown error');
        return false;
    }
    if ($code === 200 && $state === 'running') {
        $scanRunId = $_GET['scan_id'];
        return true;
    }
    $scanMessage = 'Connection scan not found (it may have expired). Run it again.';
    return false;
}

function do_import_connections(&$importedIPs, &$scanMessage) {
    // STUB: the full port-editor + edge-import flow lives in the legacy
    // connections.php today (see legacy render_port_editor + do_import
    // in frontend/php/action/connections.php). Phase 4 of the webui
    // integration plan moves it here. For now, a link to connections.php
    // gives the operator the full flow.
    $scanMessage = 'Connection import lives in the legacy page for now. <a href="connections.php">Open legacy</a>.';
}

// --- panels -----------------------------------------------------------------

function scan_connections_live_panel_html($genericProfiles, $scanMessage) {
    $f = scan_connections_form_state();
    ob_start();
    ?>
    <div class="box">
        <h3>Live connection scan</h3>
        <p style="color:#666; font-size:0.9em;">Discover LLDP / CDP / FDB edges across a set of devices or a subnet.</p>
        <form method="post" action="import.php" enctype="multipart/form-data">
            <label>Targets:
                <select name="mode">
                    <option value="from-db"   <?= $f['mode']==='from-db'?'selected':'' ?>>DB devices (mgmt IP + profile)</option>
                    <option value="profiles" <?= $f['mode']==='profiles'?'selected':'' ?>>All scan profiles</option>
                    <option value="subnet"    <?= $f['mode']==='subnet'?'selected':'' ?>>Subnet sweep</option>
                </select>
            </label><br>
            <label>Subnet (CIDR; comma-separate several):
                <input type="text" name="subnet" value="<?= htmlspecialchars($f['subnet']) ?>" placeholder="10.0.0.0/24, 10.0.1.0/24" size="40">
            </label><br>
            <label>SNMP community:
                <input type="text" name="community" value="<?= htmlspecialchars($f['community']) ?>">
            </label><br>
            <label>Collector (blank = all sources):
                <select name="collector">
                    <option value="">all (merged)</option>
                    <?php foreach (['snmp-lldp','snmp-cdp','snmp-fdb','ssh-lldp','ssh-fdb'] as $c): ?>
                        <option value="<?= $c ?>" <?= $f['collector']===$c?'selected':'' ?>><?= $c ?></option>
                    <?php endforeach; ?>
                </select>
            </label><br>

            <details<?= $f['mode']==='subnet'?' open':'' ?>><summary>Runtime SSH (subnet mode — collect LLDP/FDB from SSH-reachable hosts)</summary>
                <label>SSH user: <input type="text" name="ssh_user" value="<?= htmlspecialchars($f['ssh_user']) ?>" placeholder="root"></label>
                <label>SSH key file (on the API host): <input type="text" name="ssh_key" value="<?= htmlspecialchars($f['ssh_key']) ?>" placeholder="/home/.../.ssh/id_ed25519"></label>
                <label>SSH password: <input type="password" name="ssh_password" value="<?= htmlspecialchars($f['ssh_password']) ?>"></label>
            </details>
            <details><summary>Bring SSH credentials at runtime (uploaded keys are held in memory only)</summary>
                <label>Generic profile (reusable SSH credentials):
                    <select name="generic_profile">
                        <option value="">— none —</option>
                        <?php foreach (($genericProfiles ?? []) as $gp): $gn = $gp['name'] ?? ''; ?>
                            <option value="<?= htmlspecialchars($gn) ?>" <?= $f['generic_profile']===$gn?'selected':'' ?>><?= htmlspecialchars($gn) ?></option>
                        <?php endforeach; ?>
                    </select>
                </label><br>
                <p style="margin:4px 0; color:#777; font-size:0.85em;">Encrypted SSH secrets stored in a profile are decrypted by the credential vault — unlock it from the app bar before scanning.</p>
                <label>OpenSSH config file: <input type="file" name="ssh_config"></label><br>
                <label>SSH key file(s) referenced by the config: <input type="file" name="ssh_keys[]" multiple></label>
                <p style="margin:4px 0; color:#777; font-size:0.85em;">Keys are matched to the config by file basename, so each uploaded key must have a unique name.</p>
            </details>

            <label>Per-host SNMP timeout (s):
                <input type="number" name="timeout" value="<?= htmlspecialchars($f['timeout']) ?>" min="1" max="60" style="width:60px;">
            </label><br>
            <button type="submit" name="do_scan" value="1">Discover</button>
        </form>
        <?php if ($scanMessage): ?>
            <p style="color:#b00;"><?= $scanMessage ?></p>
        <?php endif; ?>
    </div>
    <?php
    return ob_get_clean();
}

function scan_connections_results_panel_html($result, $scanMessage) {
    if (!is_array($result) || empty($result)) {
        return '';
    }
    $hosts    = $result['hosts'] ?? [];
    $edges    = $result['edges'] ?? [];
    $inter    = $result['intermediaries'] ?? [];
    $allDisc  = $result['discrepancies'] ?? [];

    // Inline topology diagram (re-uses the same diagram endpoint the legacy
    // page hits). Returns an SVG string.
    $dgSvg = '';
    list($dgCode, $dgBody) = api_post_json(SCAN_CONNECTIONS_DIAGRAM_ENDPOINT, json_encode($result), 30);
    if ($dgCode === 200 && $dgBody !== '') {
        $dgSvg = $dgBody;
    }

    ob_start();
    ?>
    <section class="connections-results">
      <h3>Connection scan results</h3>
      <p style="color:#555; font-size:0.9em;">
        Hosts discovered: <strong><?= count($hosts) ?></strong>.
        Edges: <strong><?= count($edges) ?></strong><?php if ($inter): ?>.
        Intermediary links: <strong><?= count($inter) ?></strong><?php endif; ?>.
      </p>

      <?php if ($dgSvg !== ''): ?>
        <details open>
          <summary>Topology diagram</summary>
          <div class="resizable-img-container" style="height:480px; border:1px solid #ddd; overflow:auto; margin:8px 0;">
            <?= $dgSvg ?>
          </div>
        </details>
      <?php endif; ?>
            <button type="submit" name="do_scan_connections" value="1">Discover</button>
      <?php if (!empty($edges)): ?>
        <details open>
          <summary>Discovered edges (<?= count($edges) ?>)</summary>
          <table border="1" cellpadding="3" cellspacing="0">
            <thead><tr><th>Mark</th><th>From</th><th>To</th><th>Via</th></tr></thead>
            <tbody>
              <?php foreach ($edges as $e):
                $mark = $e['confidence'] ?? ($e['remote_resolved'] ? 'confirmed' : 'unresolved');
                $via  = implode(', ', $e['provenance'] ?? []);
              ?>
                <tr>
                  <td><?= htmlspecialchars($mark) ?></td>
                  <td><?= htmlspecialchars($e['from'] ?? '') ?></td>
                  <td><?= htmlspecialchars($e['to'] ?? '') ?></td>
                  <td style="color:#555; font-size:0.85em;"><?= htmlspecialchars($via) ?></td>
                </tr>
              <?php endforeach; ?>
            </tbody>
          </table>
        </details>
      <?php endif; ?>

      <?php if (!empty($inter)): ?>
        <details>
          <summary>Intermediary links (<?= count($inter) ?>)</summary>
          <table border="1" cellpadding="3" cellspacing="0">
            <thead><tr><th>Placeholder</th><th>Seen by</th></tr></thead>
            <tbody>
              <?php foreach ($inter as $in): ?>
                <tr>
                  <td><?= htmlspecialchars($in['device'] ?? '') ?></td>
                  <td><?= htmlspecialchars(implode(', ', $in['seen_by'] ?? [])) ?></td>
                </tr>
              <?php endforeach; ?>
            </tbody>
          </table>
        </details>
      <?php endif; ?>

      <?php if ($scanMessage): ?>
        <p><em><?= $scanMessage ?></em></p>
      <?php endif; ?>

      <p style="color:#777; font-size:0.85em;">
        Edge import + port-edit flow is in the legacy page for now:
        <a href="connections.php">Open legacy</a>.
      </p>
    </section>
    <?php
    return ob_get_clean();
}