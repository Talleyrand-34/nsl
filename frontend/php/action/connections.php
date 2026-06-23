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
    $add = function ($a, $ap, $b, $bp, $tag, $lldp, $intermediary) use (&$adj, &$degree, &$hasLLDP, &$realEdge) {
        $adj[$a][] = ['peer' => $b, 'lp' => $ap, 'pp' => $bp, 'tag' => $tag];
        $adj[$b][] = ['peer' => $a, 'lp' => $bp, 'pp' => $ap, 'tag' => $tag];
        $degree[$a] = ($degree[$a] ?? 0) + 1;
        $degree[$b] = ($degree[$b] ?? 0) + 1;
        if ($lldp) { $hasLLDP[$a] = true; $hasLLDP[$b] = true; }
        if (!$intermediary) { $realEdge[$a] = true; $realEdge[$b] = true; }
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
            list($d) = $split($sb);
            if ($d !== $hub && ($degree[$d] ?? 0) === 0) $add($hub, $hubPort, $d, '', $tag, false, true);
        }
    }
    if (!$adj) return '';

    $nodes = array_keys($adj); sort($nodes);
    $root = $nodes[0]; $bestScore = PHP_INT_MAX;
    foreach ($nodes as $n) {
        if (empty($realEdge[$n])) continue;
        $score = ($degree[$n] ?? 0) + (!empty($hasLLDP[$n]) ? 1000 : 0);
        if ($score < $bestScore) { $bestScore = $score; $root = $n; }
    }

    $visited = [$root => true]; $children = [];
    $order = function (&$links) use ($degree) {
        usort($links, function ($a, $b) use ($degree) {
            $da = $degree[$a['peer']] ?? 0; $db = $degree[$b['peer']] ?? 0;
            if ($da !== $db) return $da - $db;
            if ($a['lp'] !== $b['lp']) return strcmp($a['lp'], $b['lp']);
            return strcmp($a['peer'], $b['peer']);
        });
    };
    $build = function ($dev) use (&$build, &$visited, &$children, &$adj, $order) {
        $links = $adj[$dev] ?? []; $order($links);
        foreach ($links as $l) {
            if (!empty($visited[$l['peer']])) continue;
            $visited[$l['peer']] = true;
            $children[$dev][] = $l;
            $build($l['peer']);
        }
    };
    $build($root);

    $out = $root . "\n";
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
    $render($root, '');
    return $out;
}

// ---------------------------------------------------------------------------

$scanMessage = '';
$importMessage = '';
$result = null;

$f = [
    'mode'       => $_POST['mode'] ?? 'from-db',
    'subnet'     => $_POST['subnet'] ?? '',
    'community'  => $_POST['community'] ?? 'public',
    'collector'  => $_POST['collector'] ?? '',
    'passphrase' => $_POST['passphrase'] ?? '',
    'timeout'    => $_POST['timeout'] ?? '10',
];

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_scan'])) {
    $opts = [
        'from_db'     => $f['mode'] === 'from-db',
        'profiles'    => $f['mode'] === 'profiles',
        'subnet'      => $f['mode'] === 'subnet' ? trim($f['subnet']) : '',
        'community'   => $f['community'],
        'collector'   => $f['collector'],
        'timeout_sec' => intval($f['timeout']),
        'passphrase'  => $f['passphrase'],
    ];
    list($code, $body, $err) = api_post_json_conn(SCAN_CONNECTIONS_ENDPOINT, json_encode($opts), max(60, intval($f["timeout"]) * 8));
    if ($code === 200) {
        $result = json_decode($body, true);
    } else {
        $scanMessage = 'Scan failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body . ' ' . $err);
    }
}

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_import'])) {
    $all = json_decode($_POST['result_json'] ?? '[]', true);
    $edges = is_array($all) ? ($all['edges'] ?? []) : [];
    $sel = $_POST['edge'] ?? [];
    $toImport = [];
    foreach ($sel as $i) {
        $i = intval($i);
        if (isset($edges[$i]) && edge_importable($edges[$i])) $toImport[] = $edges[$i];
    }
    if ($toImport) {
        list($code, $body) = api_post_json_conn(SCAN_CONNECTIONS_IMPORT_ENDPOINT, json_encode(['edges' => $toImport]));
        $resp = json_decode($body, true);
        $importMessage = 'Imported ' . intval($resp['imported'] ?? 0) . ' connection(s).'
            . (!empty($resp['message']) ? ' Note: ' . htmlspecialchars($resp['message']) : '');
    } else {
        $importMessage = 'No importable edges were selected.';
    }
    $result = $all; // keep showing the result after import
}
?>

<h2>Scan connections (discover L2/L1 links)</h2>
<p style="max-width: 70ch; color:#444;">
  Gathers LLDP / CDP / bridge-FDB evidence from your hosts, correlates it into
  connection edges (and detects switches "in the middle"), and lets you import
  the ones you choose. The tool only reads — it never configures the targets.
</p>

<form method="post" class="box" style="max-width: 640px;">
    <h3>Run discovery</h3>
    <label>Targets:
        <select name="mode">
            <option value="from-db"  <?= $f['mode']==='from-db'?'selected':'' ?>>DB devices (mgmt IP + profile)</option>
            <option value="profiles" <?= $f['mode']==='profiles'?'selected':'' ?>>All scan profiles</option>
            <option value="subnet"   <?= $f['mode']==='subnet'?'selected':'' ?>>Subnet sweep</option>
        </select>
    </label><br>
    <label>Subnet (CIDR, for subnet mode):
        <input type="text" name="subnet" value="<?= htmlspecialchars($f['subnet']) ?>" placeholder="10.0.0.0/24">
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
    <label>SSH passphrase (only if a profile stores an SSH password):
        <input type="password" name="passphrase" value="<?= htmlspecialchars($f['passphrase']) ?>">
    </label><br>
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

<?php if (is_array($result)): ?>
    <?php $hosts = $result['hosts'] ?? []; $edges = $result['edges'] ?? []; $inter = $result['intermediaries'] ?? []; $disc = $result['discrepancies'] ?? []; ?>

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
        <ul>
        <?php foreach ($inter as $in): ?>
            <li><b><?= htmlspecialchars($in['mac'] ?? '') ?></b> (<?= htmlspecialchars($in['vendor'] ?? '') ?>)
                — seen by <?= htmlspecialchars(implode(', ', $in['seen_by'] ?? [])) ?></li>
        <?php endforeach; ?>
        </ul>
    <?php endif; ?>

    <?php $tree = render_topology($result); if ($tree !== ''): ?>
        <h3>Discovered topology</h3>
        <pre style="background:#f4f4f4; border:1px solid #ddd; padding:10px; overflow:auto;"><?= htmlspecialchars($tree) ?></pre>
    <?php endif; ?>

    <h3>Derived edges (<?= count($edges) ?>)</h3>
    <?php if ($edges): ?>
    <form method="post">
        <input type="hidden" name="result_json" value="<?= htmlspecialchars(json_encode($result)) ?>">
        <table border="1" cellpadding="4">
            <tr><th>Import</th><th>Mark</th><th>From</th><th>To</th><th>Via</th></tr>
            <?php foreach ($edges as $i => $e):
                $mark = edge_mark($e);
                $imp  = edge_importable($e);
                $checked = ($imp && in_array($mark, ['confirmed','candidate'])) ? 'checked' : ''; ?>
                <tr>
                    <td style="text-align:center;">
                        <?php if ($imp): ?>
                            <input type="checkbox" name="edge[]" value="<?= $i ?>" <?= $checked ?>>
                        <?php else: ?>
                            <span title="not importable">—</span>
                        <?php endif; ?>
                    </td>
                    <td><?= htmlspecialchars($mark) ?><?php if ($imp && edge_needs_create($e)): ?><br><small style="color:#a60;">(creates port)</small><?php endif; ?></td>
                    <td><?= htmlspecialchars($e['from'] ?? '') ?></td>
                    <td><?= htmlspecialchars($e['to'] ?? '') ?></td>
                    <td style="font-size:90%; color:#555;"><?= htmlspecialchars(implode(', ', $e['provenance'] ?? [])) ?></td>
                </tr>
            <?php endforeach; ?>
        </table>
        <p><button type="submit" name="do_import" value="1">Import selected connections</button></p>
    </form>
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
