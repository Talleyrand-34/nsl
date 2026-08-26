<?php
/**
 * push.php — push panel for `nsl-graph`.
 *
 * Lists devices in a selected zone with last-push timestamp and a Push button.
 * OpenWrt devices have an enabled button; other vendors get a disabled button
 * with a tooltip pointing at docs/push.md.
 *
 * The push itself never happens client-side: the button hits the backend
 * endpoint /push/device (POST {device_id, apply: bool}) which runs
 * `nsl-graph push device --device <id>` server-side. SSH credentials stay on
 * the server; the frontend never sees them.
 */

require_once 'config.php';
require_once 'header.php';

$zone = $_GET['zone'] ?? '';
$supported = ['openwrt'];
$devices = $api->get_devices_in_zone($zone);

?>
<h2>Push &mdash; zone <?php echo htmlspecialchars($zone); ?></h2>

<table class="push-table">
  <thead>
    <tr>
      <th>Device</th>
      <th>OS</th>
      <th>Last push</th>
      <th>Action</th>
    </tr>
  </thead>
  <tbody>
    <?php foreach ($devices as $dev): ?>
      <?php $os = $dev['os_type']; ?>
      <?php $supported = in_array($os, $supported, true); ?>
      <tr>
        <td><?php echo htmlspecialchars($dev['label']); ?></td>
        <td><?php echo htmlspecialchars($os); ?></td>
        <td><?php echo htmlspecialchars($dev['last_push_at'] ?? 'never'); ?></td>
        <td>
          <button
            class="push-btn"
            data-device-id="<?php echo htmlspecialchars($dev['id']); ?>"
            <?php if (!$supported): ?>disabled title="Renderer not registered; see docs/push.md"<?php endif; ?>>
            Push
          </button>
        </td>
      </tr>
    <?php endforeach; ?>
  </tbody>
</table>

<script src="push.js"></script>

<?php require_once 'footer.php'; ?>
