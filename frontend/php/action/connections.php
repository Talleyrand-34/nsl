<?php
require_once __DIR__ . '/../config.php';

// Connection discovery can scan many hosts; don't let PHP kill the request.
@set_time_limit(300);

/** api_post_json POSTs a JSON string; returns [httpCode, body, curlError]. */
if (!function_exists('api_post_json_conn')) {
    function api_post_json_conn($url, $jsonBody, $timeoutSec = 280) {
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

/** api_get_conn GETs a URL; returns [httpCode, body]. */
if (!function_exists('api_get_conn')) {
    function api_get_conn($url, $timeoutSec = 15) {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $timeoutSec);
        $body = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        return [$code, $body];
    }
}

/** edge_mark mirrors the CLI: confirmed/candidate/weak, else possible/unresolved. */
function edge_mark($e) {
    if (!empty($e['remote_resolved'])) {
        return $e['confidence'] ?? '';
    }
    return (strpos($e['to'] ?? '', 'unknown(') === 0) ? 'unresolved' : 'possible';
}

/** endpoint_port returns the port name from a "device:port" label ('' if device-level/unknown). */
function endpoint_port($label) {
    if (strpos($label, 'unknown(') === 0) return '';
    $i = strrpos($label, ':');
    return ($i !== false && $i + 1 < strlen($label)) ? substr($label, $i + 1) : '';
}

/** edge_pair_key builds an order-independent key for an undirected "a"/"b" pair. */
function edge_pair_key($a, $b) {
    return $a < $b ? $a . "\x00" . $b : $b . "\x00" . $a;
}

/** edge_in_db reports whether a derived edge already exists as a DB connection. */
function edge_in_db($e, $dbPairs) {
    return isset($dbPairs[edge_pair_key($e['from'] ?? '', $e['to'] ?? '')]);
}

/** endpoint_device returns the device part of a "device:port" label ('' if unknown). */
function endpoint_device($label) {
    if ($label === '' || strpos($label, 'unknown(') === 0) return '';
    $i = strrpos($label, ':');
    return ($i !== false) ? substr($label, 0, $i) : $label;
}

/**
 * render_port_editor prints the editable port chooser for one edge endpoint.
 * Options, in priority order: the detected port (default), enter-manually, the
 * device's existing device ports, then its model's model ports. Empty device
 * (unknown endpoint) just shows the label.
 */
function render_port_editor($side, $i, $label, $modelByDevice, $portsByDevice, $modelPortsByModel) {
    $dev = endpoint_device($label);
    $detected = endpoint_port($label);
    if ($dev === '') {
        echo htmlspecialchars($label); // unknown / unresolvable endpoint
        return;
    }
    $dps = $portsByDevice[$dev] ?? [];
    $mps = $modelPortsByModel[$modelByDevice[$dev] ?? ''] ?? [];
    $seen = [];
    echo htmlspecialchars($dev) . ':';
    echo '<select name="' . $side . '_port[' . $i . ']" onchange="nslPortManual(this)">';
    if ($detected !== '') {
        echo '<option value="' . htmlspecialchars($detected) . '" selected>detected: ' . htmlspecialchars($detected) . '</option>';
        $seen[$detected] = true;
    }
    echo '<option value="__manual__"' . ($detected === '' ? ' selected' : '') . '>&#9998; enter manually&hellip;</option>';
    if ($dps) {
        echo '<optgroup label="Device ports">';
        foreach ($dps as $p) {
            if (isset($seen[$p])) continue;
            $seen[$p] = true;
            echo '<option value="' . htmlspecialchars($p) . '">' . htmlspecialchars($p) . '</option>';
        }
        echo '</optgroup>';
    }
    if ($mps) {
        echo '<optgroup label="Model ports">';
        foreach ($mps as $p) {
            if (isset($seen[$p])) continue;
            $seen[$p] = true;
            echo '<option value="' . htmlspecialchars($p) . '">' . htmlspecialchars($p) . '</option>';
        }
        echo '</optgroup>';
    }
    echo '</select>';
    echo '<input type="text" name="' . $side . '_port_manual[' . $i . ']" placeholder="port name" style="' .
        ($detected === '' ? '' : 'display:none;') . 'width:9em;">';
}

/** edge_importable: both endpoints name a port (existing or creatable on import). */
function edge_importable($e) {
    $hasFrom = !empty($e['from_deviceport_id']) || endpoint_port($e['from'] ?? '') !== '';
    $hasTo   = !empty($e['to_deviceport_id'])   || endpoint_port($e['to'] ?? '') !== '';
    return $hasFrom && $hasTo;
}

/** edge_needs_create: a missing device port will be created on import. */
function edge_needs_create($e) {
    return empty($e['from_deviceport_id']) || empty($e['to_deviceport_id']);
}

/** render_topology builds the ASCII tree from the discovered edges (port of the CLI). */
function render_topology($result) {
    $edges = $result['edges'] ?? [];
    $inter = $result['intermediaries'] ?? [];
    if (!$edges && !$inter) return '';

    $split = function ($s) {
        if (strpos($s, 'unknown(') === 0) return [$s, ''];
        $i = strrpos($s, ':');
        if ($i !== false) return [substr($s, 0, $i), substr($s, $i + 1)];
        return [$s, ''];
    };
    $adj = []; $degree = []; $hasLLDP = []; $realEdge = [];
    $allEdges = [];
    $add = function ($a, $ap, $b, $bp, $tag, $lldp, $intermediary) use (&$adj, &$degree, &$hasLLDP, &$realEdge, &$allEdges) {
        $adj[$a][] = ['peer' => $b, 'lp' => $ap, 'pp' => $bp, 'tag' => $tag];
        $adj[$b][] = ['peer' => $a, 'lp' => $bp, 'pp' => $ap, 'tag' => $tag];
        $degree[$a] = ($degree[$a] ?? 0) + 1;
        $degree[$b] = ($degree[$b] ?? 0) + 1;
        if ($lldp) { $hasLLDP[$a] = true; $hasLLDP[$b] = true; }
        if (!$intermediary) { $realEdge[$a] = true; $realEdge[$b] = true; }
        $allEdges[] = ['a' => $a, 'ap' => $ap, 'b' => $b, 'bp' => $bp, 'tag' => $tag];
    };
    $srcOf = function ($provs) {
        $set = [];
        foreach (($provs ?? []) as $p) {
            $s = $p; $at = strpos($p, '@'); if ($at !== false) $s = substr($p, 0, $at);
            $set[$s] = true;
        }
        return implode('+', array_keys($set));
    };

    foreach ($edges as $e) {
        list($aD, $aP) = $split($e['from'] ?? '');
        list($bD, $bP) = $split($e['to'] ?? '');
        if ($aD === '' || $bD === '' || $aD === $bD) continue;
        $src = $srcOf($e['provenance'] ?? []);
        $tag = edge_mark($e) . ($src ? ' ' . $src : '');
        $add($aD, $aP, $bD, $bP, $tag, strpos($src, 'lldp') !== false, false);
    }
    foreach ($inter as $in) {
        $hub = ''; $hubPort = ''; $best = -1;
        foreach (($in['seen_by'] ?? []) as $sb) {
            list($d, $p) = $split($sb);
            if (($degree[$d] ?? 0) > $best) { $best = $degree[$d] ?? 0; $hub = $d; $hubPort = $p; }
        }
        if ($hub === '') continue;
        $tag = 'via ' . ($in['vendor'] ?? 'switch') . ' ' . ($in['mac'] ?? '');
        foreach (($in['seen_by'] ?? []) as $sb) {
            list($d, $p) = $split($sb);
            if ($d !== $hub) $add($hub, $hubPort, $d, $p, $tag, false, true);
        }
    }
    if (!$adj) return '';

    // Roots are tried best-first (prefer real LLDP endpoints, leafs first) so the
    // main fabric reads top-down; every component is rendered, not just one.
    $nodes = array_keys($adj); sort($nodes);
    usort($nodes, function ($a, $b) use ($degree, $hasLLDP, $realEdge) {
        $sa = ($degree[$a] ?? 0) + (!empty($hasLLDP[$a]) ? 1000 : 0) + (empty($realEdge[$a]) ? 100000 : 0);
        $sb = ($degree[$b] ?? 0) + (!empty($hasLLDP[$b]) ? 1000 : 0) + (empty($realEdge[$b]) ? 100000 : 0);
        if ($sa !== $sb) return $sa - $sb;
        return strcmp($a, $b);
    });

    $order = function (&$links) use ($degree) {
        usort($links, function ($a, $b) use ($degree) {
            $da = $degree[$a['peer']] ?? 0; $db = $degree[$b['peer']] ?? 0;
            if ($da !== $db) return $da - $db;
            if ($a['lp'] !== $b['lp']) return strcmp($a['lp'], $b['lp']);
            return strcmp($a['peer'], $b['peer']);
        });
    };

    // ekey is an order-independent key for an undirected device pair.
    $ekey = function ($x, $y) { return $x < $y ? $x . "\x00" . $y : $y . "\x00" . $x; };

    $visited = [];
    $treeEdge = [];   // ekey -> true for edges drawn as part of a tree
    $children = [];
    $build = function ($dev) use (&$build, &$visited, &$children, &$adj, $order, $ekey, &$treeEdge) {
        $links = $adj[$dev] ?? []; $order($links);
        foreach ($links as $l) {
            if (!empty($visited[$l['peer']])) continue;
            $visited[$l['peer']] = true;
            $treeEdge[$ekey($dev, $l['peer'])] = true;
            $children[$dev][] = $l;
            $build($l['peer']);
        }
    };

    $out = '';
    $render = function ($dev, $prefix) use (&$render, &$children, &$out) {
        $kids = $children[$dev] ?? [];
        $n = count($kids);
        foreach ($kids as $i => $l) {
            $last = ($i === $n - 1);
            $branch = $last ? '└── ' : '├── ';
            $cp = $prefix . ($last ? '    ' : '│   ');
            $lp = $l['lp'] !== '' ? $l['lp'] : '·';
            $peer = $l['peer'] . ($l['pp'] !== '' ? ':' . $l['pp'] : '');
            $out .= $prefix . $branch . $lp . ' ─[' . $l['tag'] . ']─ ' . $peer . "\n";
            $render($l['peer'], $cp);
        }
    };

    // Render every connected component as its own tree.
    foreach ($nodes as $root) {
        if (!empty($visited[$root])) continue;
        $visited[$root] = true;
        $build($root);
        $out .= $root . "\n";
        $render($root, '');
    }

    // Any edge not drawn as a tree edge (extra links / cycles) is listed so all
    // connections are shown, not just the spanning tree.
    $extras = [];
    $seenExtra = [];
    foreach ($allEdges as $e) {
        $k = $ekey($e['a'], $e['b']);
        if (!empty($treeEdge[$k]) || !empty($seenExtra[$k])) continue;
        $seenExtra[$k] = true;
        $ap = $e['ap'] !== '' ? ':' . $e['ap'] : '';
        $bp = $e['bp'] !== '' ? ':' . $e['bp'] : '';
        $extras[] = '  ' . $e['a'] . $ap . ' ─[' . $e['tag'] . ']─ ' . $e['b'] . $bp;
    }
    if ($extras) {
        $out .= "\nAdditional links:\n" . implode("\n", $extras) . "\n";
    }
    return $out;
}

// ---------------------------------------------------------------------------

$scanMessage = '';
$importMessage = '';
$result = null;
$scanRunId = '';   // when set, a scan just started: render the live panel

// Generic profiles (reusable SSH credentials), for the fallback selector.
$allProfiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true);
if (!is_array($allProfiles)) {
    $allProfiles = [];
}
$genericProfiles = array_values(array_filter($allProfiles, fn($p) => ($p['kind'] ?? '') === 'generic'));

$f = [
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

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_scan'])) {
    // Uploaded OpenSSH config + key files are read into memory and sent in the
    // request body — never written to disk on this host.
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
    } else {
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
        list($code, $body, $err) = api_post_json_conn(SCAN_CONNECTIONS_ENDPOINT, json_encode($opts), 30);
        $j = json_decode($body, true);
        if (($code === 202 || $code === 200) && !empty($j['scan_id'])) {
            $scanRunId = $j['scan_id'];
        } else {
            $scanMessage = 'Could not start scan (HTTP ' . intval($code) . '): ' . htmlspecialchars($body . ' ' . $err);
        }
    }
}

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_import'])) {
    $all = json_decode($_POST['result_json'] ?? '[]', true);
    $edges = is_array($all) ? ($all['edges'] ?? []) : [];
    $sel = $_POST['edge'] ?? [];

    // Resolve the port the user chose for one endpoint of an edge. Priority:
    // an explicit manual entry, else the selected value (which defaults to the
    // detected port, or lists existing device/model ports). Empty = unchanged.
    $chosenPort = function ($side, $i) {
        $v = $_POST[$side . '_port'][$i] ?? '';
        if ($v === '__manual__') {
            $v = trim($_POST[$side . '_port_manual'][$i] ?? '');
        }
        return $v;
    };
    $toImport = [];
    foreach ($sel as $i) {
        $i = intval($i);
        if (!isset($edges[$i])) continue;
        $edge = $edges[$i];

        // Apply any port edits: rebuild "device:port" and drop the resolved
        // deviceport id so the backend re-resolves (or creates) the chosen port.
        foreach (['from', 'to'] as $side) {
            $dev = endpoint_device($edge[$side] ?? '');
            $port = $chosenPort($side, $i);
            if ($dev !== '' && $port !== '') {
                $edge[$side] = $dev . ':' . $port;
                unset($edge[$side . '_deviceport_id']);
            }
        }

        if (edge_importable($edge)) $toImport[] = $edge;
    }
    if ($toImport) {
        list($code, $body) = api_post_json_conn(SCAN_CONNECTIONS_IMPORT_ENDPOINT, json_encode(['edges' => $toImport]));
        $resp = json_decode($body, true);
        $importMessage = 'Imported ' . intval($resp['imported'] ?? 0) . ' connection(s).'
            . (!empty($resp['message']) ? ' Note: ' . htmlspecialchars($resp['message']) : '');
    }

    // Intermediary ("via the middle device") links selected in the edges table
    // are materialized through the shared placeholder device, connecting only the
    // observing endpoints the user picked (values are "<intermediaryIdx>_<seenByIdx>").
    $interAll = is_array($all) ? ($all['intermediaries'] ?? []) : [];
    $byInter = [];
    foreach (($_POST['inter_edge'] ?? []) as $v) {
        $parts = explode('_', $v);
        if (count($parts) !== 2) continue;
        list($j, $k) = [intval($parts[0]), intval($parts[1])];
        if (isset($interAll[$j]['seen_by'][$k])) $byInter[$j][] = $interAll[$j]['seen_by'][$k];
    }
    if ($byInter) {
        $interPayload = [];
        foreach ($byInter as $j => $endpoints) {
            $im = $interAll[$j];
            $im['seen_by'] = array_values(array_unique($endpoints));
            $interPayload[] = $im;
        }
        list($pc, $pb) = api_post_json_conn(SCAN_CONNECTIONS_PLACEHOLDERS_ENDPOINT, json_encode(['intermediaries' => $interPayload]));
        $presp = json_decode($pb, true);
        $importMessage .= ($importMessage ? ' ' : '') . 'Placeholder "' . htmlspecialchars($presp['device'] ?? '?') . '": '
            . intval($presp['connections'] ?? 0) . ' connection(s).'
            . (!empty($presp['message']) ? ' Note: ' . htmlspecialchars($presp['message']) : '');
    }

    if ($importMessage === '') {
        $importMessage = 'No connections were selected.';
    }
    $result = $all; // keep showing the result after import
}

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_placeholders'])) {
    $all = json_decode($_POST['result_json'] ?? '[]', true);
    $inter = is_array($all) ? ($all['intermediaries'] ?? []) : [];
    $sel = $_POST['inter'] ?? [];
    $toCreate = [];
    foreach ($sel as $i) {
        $i = intval($i);
        if (isset($inter[$i])) $toCreate[] = $inter[$i];
    }
    if ($toCreate) {
        list($code, $body) = api_post_json_conn(SCAN_CONNECTIONS_PLACEHOLDERS_ENDPOINT, json_encode(['intermediaries' => $toCreate]));
        $resp = json_decode($body, true);
        $importMessage = 'Created placeholder device "' . htmlspecialchars($resp['device'] ?? '?')
            . '" in zone "' . htmlspecialchars($resp['zone'] ?? '?') . '" with '
            . intval($resp['connections'] ?? 0) . ' connection(s).'
            . (!empty($resp['message']) ? ' Note: ' . htmlspecialchars($resp['message']) : '');
    } else {
        $importMessage = 'No intermediaries were selected.';
    }
    $result = $all; // keep showing the result after creating placeholders
}

// --- A started scan finished — fetch its result by id and render -------------
if ($_SERVER['REQUEST_METHOD'] === 'GET' && isset($_GET['scan_id'])) {
    list($code, $body) = api_get_conn(SCAN_STATUS_ENDPOINT . '?scan_id=' . urlencode($_GET['scan_id']));
    $st = json_decode($body, true);
    $state = is_array($st) ? ($st['state'] ?? '') : '';
    if ($code === 200 && $state === 'completed') {
        $result = $st['result'] ?? [];
    } elseif ($code === 200 && $state === 'failed') {
        $scanMessage = 'Scan failed: ' . htmlspecialchars($st['error'] ?? 'unknown error');
    } elseif ($code === 200 && $state === 'running') {
        $scanRunId = $_GET['scan_id'];
    } else {
        $scanMessage = 'Scan not found (it may have expired). Run it again.';
    }
}
?>

<h2>Scan connections (discover L2/L1 links)</h2>
<p style="max-width: 70ch; color:#444;">
  Gathers LLDP / CDP / bridge-FDB evidence from your hosts, correlates it into
  connection edges (and detects switches "in the middle"), and lets you import
  the ones you choose. The tool only reads — it never configures the targets.
</p>

<form method="post" class="box" style="max-width: 640px;" enctype="multipart/form-data">
    <h3>Run discovery</h3>
    <label>Targets:
        <select name="mode">
            <option value="from-db"  <?= $f['mode']==='from-db'?'selected':'' ?>>DB devices (mgmt IP + profile)</option>
            <option value="profiles" <?= $f['mode']==='profiles'?'selected':'' ?>>All scan profiles</option>
            <option value="subnet"   <?= $f['mode']==='subnet'?'selected':'' ?>>Subnet sweep</option>
        </select>
    </label><br>
    <label>Subnet (CIDR, for subnet mode — comma-separate several):
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
    <fieldset style="margin:6px 0; border:1px solid #ddd;">
        <legend style="font-size:90%;">Runtime SSH (subnet mode — collect LLDP/FDB from SSH-reachable hosts, no profile needed)</legend>
        <label>SSH user: <input type="text" name="ssh_user" value="<?= htmlspecialchars($f['ssh_user']) ?>" placeholder="root"></label>
        <label>SSH key file (on the API host): <input type="text" name="ssh_key" value="<?= htmlspecialchars($f['ssh_key']) ?>" placeholder="/home/.../.ssh/id_ed25519"></label>
        <label>SSH password: <input type="password" name="ssh_password" value="<?= htmlspecialchars($f['ssh_password']) ?>"></label>
    </fieldset>
    <fieldset style="margin:6px 0; border:1px solid #ddd;">
        <legend style="font-size:90%;">Bring SSH credentials at runtime (uploaded keys are held in memory only, never stored)</legend>
        <label>Generic profile (reusable SSH credentials):
            <select name="generic_profile">
                <option value="">— none —</option>
                <?php foreach ($genericProfiles as $gp): $gn = $gp['name'] ?? ''; ?>
                    <option value="<?= htmlspecialchars($gn) ?>" <?= $f['generic_profile']===$gn?'selected':'' ?>><?= htmlspecialchars($gn) ?></option>
                <?php endforeach; ?>
            </select>
        </label><br>
        <p style="margin:4px 0; color:#777; font-size:0.85em;">Encrypted SSH secrets stored in a generic / from-db / matched device profile are decrypted by the credential vault — unlock it from the app bar before scanning.</p>
        <label>OpenSSH config file: <input type="file" name="ssh_config"></label><br>
        <label>SSH key file(s) referenced by the config: <input type="file" name="ssh_keys[]" multiple></label>
        <p style="margin:4px 0; color:#777; font-size:0.85em;">Keys are matched to the config by file basename, so each uploaded key must have a unique name.</p>
    </fieldset>
    <label>Per-host SNMP timeout (s):
        <input type="number" name="timeout" value="<?= htmlspecialchars($f['timeout']) ?>" min="1" max="60" style="width:60px;">
    </label><br>
    <button type="submit" name="do_scan" value="1">Discover</button>
</form>

<?php if ($scanMessage): ?>
    <p style="color:#b00;"><?= $scanMessage ?></p>
<?php endif; ?>
<?php if ($importMessage): ?>
    <p style="color:#070; font-weight:bold;"><?= $importMessage ?></p>
<?php endif; ?>

<?php if ($scanRunId): ?>
    <div id="scan-status" class="scan-status">
        <h4><span class="spinner"></span>Discovering connections…</h4>
        <div class="scan-state" id="scan-state">starting…</div>
        <div class="scan-bar" id="scan-bar" style="display:none;"><div class="scan-bar-fill" id="scan-bar-fill"></div></div>
        <div class="scan-events" id="scan-events"></div>
    </div>
    <script>nslWatchScan(<?= json_encode($scanRunId) ?>, {});</script>
<?php endif; ?>

<?php if (is_array($result)): ?>
    <?php
        $hosts = $result['hosts'] ?? []; $edges = $result['edges'] ?? []; $inter = $result['intermediaries'] ?? [];
        $allDisc = $result['discrepancies'] ?? [];
        $notInDb = array_values(array_filter($allDisc, fn($d) => ($d['kind'] ?? '') === 'host-not-in-db'));
        $disc    = array_values(array_filter($allDisc, fn($d) => ($d['kind'] ?? '') !== 'host-not-in-db'));
    ?>

    <?php if ($notInDb): ?>
        <div style="border:1px solid #e0a800; background:#fff8e1; padding:10px; margin-top:12px;">
            <b>⚠ Hosts discovered that are not in the DB</b> — add them first (Import devices) so their connections can resolve:
            <ul>
            <?php foreach ($notInDb as $w): ?>
                <li><?= htmlspecialchars($w['detail'] ?? '') ?></li>
            <?php endforeach; ?>
            </ul>
        </div>
    <?php endif; ?>

    <h3>Gather (<?= count($hosts) ?> host(s))</h3>
    <table border="1" cellpadding="4">
        <tr><th>Host</th><th>Device</th><th>evidence</th><th>fdb</th><th>errors</th></tr>
        <?php foreach ($hosts as $h):
            $name = $h['device_label'] ?? ''; if ($name==='' && isset($h['device']['sys_name'])) $name = $h['device']['sys_name']; ?>
            <tr>
                <td><?= htmlspecialchars($h['host'] ?? '') ?></td>
                <td><?= htmlspecialchars($name) ?></td>
                <td style="text-align:center;"><?= count($h['evidence'] ?? []) ?></td>
                <td style="text-align:center;"><?= count($h['fdb'] ?? []) ?></td>
                <td style="color:#a00;"><?= htmlspecialchars(implode('; ', $h['errors'] ?? [])) ?></td>
            </tr>
        <?php endforeach; ?>
    </table>

    <?php if ($inter): ?>
        <h3>Intermediary devices detected in the middle</h3>
        <p style="max-width:70ch; color:#555;">
          These MACs were seen in the forwarding tables of multiple hosts but don't
          speak LLDP and aren't in the DB — i.e. unknown device(s) sitting between
          known hosts. Their links to each observing host appear in <b>Derived edges</b>
          below (rows marked <i>via … (placeholder)</i>): select the ones you want and
          import to create a shared <b>placeholder unmanaged device</b> (in an
          "<?= htmlspecialchars('Unknown infrastructure') ?>" zone) connected to only
          the endpoints you pick.
        </p>
        <table border="1" cellpadding="4">
            <tr><th>MAC</th><th>Vendor</th><th>Seen by</th></tr>
            <?php foreach ($inter as $in): ?>
                <tr>
                    <td><b><?= htmlspecialchars($in['mac'] ?? '') ?></b></td>
                    <td><?= htmlspecialchars($in['vendor'] ?? '') ?></td>
                    <td style="font-size:90%; color:#555;"><?= htmlspecialchars(implode(', ', $in['seen_by'] ?? [])) ?></td>
                </tr>
            <?php endforeach; ?>
        </table>
    <?php endif; ?>

    <?php $tree = render_topology($result); if ($tree !== ''): ?>
        <h3>Discovered topology</h3>
        <?php
            // Render the discovered topology as a D2 diagram by posting the scan
            // result to the diagram endpoint (reuses the standard diagram generator).
            list($dgCode, $dgSvg) = api_post_json_conn(SCAN_CONNECTIONS_DIAGRAM_ENDPOINT, json_encode($result), 30);
        ?>
        <?php if ($dgCode === 200 && $dgSvg !== ''): ?>
            <div class="resizable-img-container" style="height:480px; border:1px solid #ddd; overflow:auto;"><?= $dgSvg ?></div>
        <?php endif; ?>
        <details<?= ($dgCode === 200 && $dgSvg !== '') ? '' : ' open' ?>>
            <summary>Text view</summary>
            <pre style="background:#f4f4f4; border:1px solid #ddd; padding:10px; overflow:auto;"><?= htmlspecialchars($tree) ?></pre>
        </details>
    <?php endif; ?>

    <h3>Derived edges (<?= count($edges) ?><?php if ($inter): ?> + intermediary links<?php endif; ?>)</h3>
    <?php if ($edges || $inter): ?>
    <?php
        // Lookups so each endpoint's port can be edited: detected -> manual ->
        // the device's device ports -> its model's model ports.
        $allDevices    = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
        $allDevPorts   = json_decode(@file_get_contents(DEVICEPORTS_ENDPOINT), true) ?: [];
        $allModelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];
        $modelByDevice = [];
        foreach ($allDevices as $d) { $modelByDevice[$d['label'] ?? ''] = $d['model'] ?? ''; }
        $portsByDevice = [];
        foreach ($allDevPorts as $p) { $portsByDevice[$p['devname'] ?? ''][] = $p['portname'] ?? ''; }
        $modelPortsByModel = [];
        foreach ($allModelPorts as $mp) { $modelPortsByModel[$mp['model'] ?? ''][] = $mp['name'] ?? ''; }
        // Existing connections, to flag: edges already in the DB ($dbPairs), ports
        // already connected ($usedPorts — a port may have only one connection), and
        // ports already wired to a placeholder ($placeholderPorts).
        $existingConns = json_decode(@file_get_contents(CONNECTIONS_ENDPOINT), true) ?: [];
        $dbPairs = [];
        $usedPorts = [];
        $placeholderPorts = [];
        $isPlaceholderEnd = function ($dev, $zone) {
            return $zone === 'Unknown infrastructure' || strpos((string) $dev, 'unmanaged') === 0;
        };
        foreach ($existingConns as $c) {
            $a = ($c['fromdevice'] ?? '') . ':' . ($c['frommodel'] ?? '');
            $b = ($c['todevice'] ?? '') . ':' . ($c['tomodel'] ?? '');
            $dbPairs[edge_pair_key($a, $b)] = true;
            $usedPorts[$a] = true;
            $usedPorts[$b] = true;
            // If one end is a placeholder, the other end's port is "mapped" to it.
            if ($isPlaceholderEnd($c['todevice'] ?? '', $c['tozonename'] ?? '')) $placeholderPorts[$a] = $c['todevice'] ?? 'placeholder';
            if ($isPlaceholderEnd($c['fromdevice'] ?? '', $c['fromzonename'] ?? '')) $placeholderPorts[$b] = $c['fromdevice'] ?? 'placeholder';
        }
    ?>
    <form method="post">
        <input type="hidden" name="result_json" value="<?= htmlspecialchars(json_encode($result)) ?>">
        <p style="color:#555; font-size:90%;">Each port can be edited: pick the detected port, enter one manually, or choose an existing device/model port.</p>
        <table border="1" cellpadding="4">
            <tr><th>Import</th><th>DB status</th><th>Mark</th><th>From</th><th>To</th><th>Via</th></tr>
            <?php foreach ($edges as $i => $e):
                $mark = edge_mark($e);
                $imp  = edge_importable($e);
                $inDb = edge_in_db($e, $dbPairs);
                // A port may hold only one connection: flag if an endpoint is already
                // in use by a different connection (importing it would be rejected).
                $portConflict = !$inDb && (isset($usedPorts[$e['from'] ?? '']) || isset($usedPorts[$e['to'] ?? '']));
                $checked = ($imp && !$inDb && !$portConflict && in_array($mark, ['confirmed','candidate'])) ? 'checked' : ''; ?>
                <tr>
                    <td style="text-align:center;">
                        <input type="checkbox" name="edge[]" value="<?= $i ?>" <?= $checked ?>>
                    </td>
                    <td style="text-align:center;">
                        <?php if ($inDb): ?><span title="this connection already exists in the database" style="color:#070;">&#10003; in DB</span>
                        <?php elseif ($portConflict): ?><span title="a port of this edge already has a connection (one connection per port)" style="color:#b00;">&#9888; port in use</span>
                        <?php endif; ?>
                    </td>
                    <td><?= htmlspecialchars($mark) ?><?php if ($imp && edge_needs_create($e)): ?><br><small style="color:#a60;">(creates port)</small><?php endif; ?></td>
                    <td><?php render_port_editor('from', $i, $e['from'] ?? '', $modelByDevice, $portsByDevice, $modelPortsByModel); ?></td>
                    <td><?php render_port_editor('to', $i, $e['to'] ?? '', $modelByDevice, $portsByDevice, $modelPortsByModel); ?></td>
                    <td style="font-size:90%; color:#555;"><?= htmlspecialchars(implode(', ', $e['provenance'] ?? [])) ?></td>
                </tr>
            <?php endforeach; ?>
            <?php // Links detected via a middle (placeholder) device — selecting one
                  // materializes the shared placeholder and connects that endpoint.
            foreach ($inter as $j => $in): foreach (($in['seen_by'] ?? []) as $k => $sb):
                // Already mapped: this observing port already connects to a placeholder.
                $mappedTo = $placeholderPorts[$sb] ?? '';
                $portUsed = $mappedTo === '' && isset($usedPorts[$sb]); ?>
                <tr style="background:#fbf7ef;">
                    <td style="text-align:center;"><input type="checkbox" name="inter_edge[]" value="<?= $j . '_' . $k ?>" <?= ($mappedTo !== '' || $portUsed) ? 'disabled' : '' ?>></td>
                    <td style="text-align:center;">
                        <?php if ($mappedTo !== ''): ?><span title="already connected to placeholder <?= htmlspecialchars($mappedTo) ?>" style="color:#070;">&#10003; mapped</span>
                        <?php elseif ($portUsed): ?><span title="this port already has a connection" style="color:#b00;">&#9888; port in use</span>
                        <?php endif; ?>
                    </td>
                    <td>via <?= htmlspecialchars($in['vendor'] ?: 'switch') ?><br><small style="color:#a60;">(placeholder)</small></td>
                    <td><?= htmlspecialchars($sb) ?></td>
                    <td>&harr; placeholder <small>(<?= htmlspecialchars($in['mac'] ?? '') ?>)</small></td>
                    <td style="font-size:90%; color:#555;">intermediary <?= htmlspecialchars($in['mac'] ?? '') ?></td>
                </tr>
            <?php endforeach; endforeach; ?>
        </table>
        <p><button type="submit" name="do_import" value="1">Import selected connections</button></p>
    </form>
    <script>
    function nslPortManual(sel) {
        var inp = sel.parentNode.querySelector('input[type=text]');
        if (!inp) return;
        inp.style.display = sel.value === '__manual__' ? '' : 'none';
    }
    </script>
    <?php else: ?>
        <p>No edges derived. Ensure LLDP/SNMP is enabled on the targets and that they have scan profiles.</p>
    <?php endif; ?>

    <?php if ($disc): ?>
        <h3>Discrepancies</h3>
        <ul>
        <?php foreach ($disc as $d): ?>
            <li>[<?= htmlspecialchars($d['kind'] ?? '') ?>] <?= htmlspecialchars($d['detail'] ?? '') ?></li>
        <?php endforeach; ?>
        </ul>
    <?php endif; ?>
<?php endif; ?>
