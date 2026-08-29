<?php
// SPDX-License-Identifier: MIT
// push_topology_panel.php — right column of push.php.
//
// Phase 6 wires the inline D2 render of the selected device + 1-hop
// neighbors. Phase 2 leaves a stub container so the layout is correct.
function push_topology_panel_html(): string
{
    ob_start(); ?>
    <div class="box">
        <h3>Topology preview</h3>
        <p style="color:#555; font-size:0.85em; margin:4px 0 12px 0;">
            Selected device and its 1-hop neighbors. Drawn from the connections
            table so you can see what your push will affect before you click Apply.
        </p>
        <div id="push-topology"
             style="min-height:320px; border:1px solid #ddd; background:#fff;
                    padding:8px; display:flex; align-items:center; justify-content:center;
                    color:#777; font-style:italic;">
            Pick a device on the left to see its topology.
        </div>
        <h4 style="margin-top:14px;">Neighbors</h4>
        <table id="push-neighbors"
               style="width:100%; border-collapse:collapse; font-size:0.85em;">
            <thead>
                <tr style="background:#f0f0f0;">
                    <th align="left">Neighbor</th>
                    <th align="left">OS</th>
                    <th align="left">Port (this → neighbor)</th>
                    <th align="left">VLAN</th>
                </tr>
            </thead>
            <tbody id="push-neighbors-body">
                <tr><td colspan="4" style="color:#777; font-style:italic;">&mdash;</td></tr>
            </tbody>
        </table>
    </div>
    <?php
    return ob_get_clean();
}
