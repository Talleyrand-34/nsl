<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// push_preview_panel.php — left column of push.php.
//
// Device picker + intent editor + four buttons (Preview, Dry-run, Apply,
// Rollback) + audit log section. Phase 5 wires the buttons to /api/v1/push/*.
function push_preview_panel_html(array $devices = []): string
{
    ob_start(); ?>
    <div class="box">
        <h3>Operations</h3>
        <p style="color:#555; font-size:0.85em; margin:4px 0 12px 0;">
            Pick a device, see its intent, and run one of the four verbs.
            Audit log below records every run.
        </p>

        <label>Device:
            <select id="push-device" name="device" style="min-width:240px;">
                <option value="">&mdash; select a device &mdash;</option>
                <?php foreach ($devices as $d):
                    $id = htmlspecialchars((string)($d['id'] ?? $d['name'] ?? ''));
                    $label = htmlspecialchars(($d['name'] ?? '(unnamed)') . ' [' . ($d['os'] ?? '?') . ']');
                    $os = htmlspecialchars((string)($d['os'] ?? ''));
                    if ($id === '') continue;
                ?>
                    <option value="<?= $id ?>" data-os="<?= $os ?>"><?= $label ?></option>
                <?php endforeach; ?>
        </label>

        <p style="margin:8px 0 4px 0;"><strong>OS:</strong> <span id="push-device-os">&mdash;</span></p>

        <label style="display:block; margin-top:6px;">Intent (read-only in v1):
            <textarea id="push-intent" rows="8" style="width:100%; font-family:monospace; font-size:0.85em;"
                      placeholder='{"hostname":"…","interfaces":[…]}' readonly></textarea>
        </label>

        <div style="margin-top:10px; display:flex; gap:6px; flex-wrap:wrap;">
            <button type="button" id="push-preview"   class="push-btn">Preview</button>
            <button type="button" id="push-dryrun"    class="push-btn">Dry-run</button>
            <button type="button" id="push-apply"     class="push-btn push-btn-danger">Apply</button>
            <button type="button" id="push-rollback"  class="push-btn">Rollback</button>
        </div>

        <div id="push-result" style="margin-top:10px;">
            <pre id="push-result-text"
                 style="display:none; background:#f6f8fa; border:1px solid #ddd; padding:8px; max-height:280px; overflow:auto; font-size:0.85em;"></pre>
        </div>

        <h4 style="margin-top:14px;">Audit log</h4>
        <div id="push-history"
             style="background:#fafafa; border:1px solid #ddd; padding:6px; max-height:240px; overflow:auto; font-size:0.85em;">
            <em style="color:#777;">No runs yet.</em>
        </div>
    </div>
    <?php
    return ob_get_clean();
}
