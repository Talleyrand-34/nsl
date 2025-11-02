
<?php
// config.php
/* define('API_BASE_URL', 'http://localhost:8081'); */
// Default API base URL
define('API_BASE_URL_DEFAULT', 'http://localhost:8081');
// Allow override via session
session_start();
define('API_BASE_URL', $_SESSION['api_base_url'] ?? API_BASE_URL_DEFAULT);
define('BRANDS_ENDPOINT', API_BASE_URL . '/brands');
define('DEVCLASSES_ENDPOINT', API_BASE_URL . '/deviceclasses');
define('PROPRIETARIES_ENDPOINT', API_BASE_URL . '/proprietaries');
define('ZONETYPES_ENDPOINT', API_BASE_URL . '/zonetypes');
define('ZONES_ENDPOINT', API_BASE_URL . '/zones');
define('MODELS_ENDPOINT', API_BASE_URL . '/models');
define('DEVICES_ENDPOINT', API_BASE_URL . '/devices');
define('MODELPORTS_ENDPOINT', API_BASE_URL . '/modelports');
define('MODELPORTS_BULK_ENDPOINT', MODELPORTS_ENDPOINT . '/bulk');
define('CONNECTIONS_ENDPOINT', API_BASE_URL . '/connections');
define('CONNECTIONTYPES_ENDPOINT', API_BASE_URL . '/connectiontypes');
define('DEVICEPORTS_ENDPOINT', API_BASE_URL . '/deviceports');
define('VLANS_ENDPOINT', API_BASE_URL . '/vlans');
define('PLUGINS_ENDPOINT', API_BASE_URL . '/plugins');
define('PLUGINS_ACTIVE_ENDPOINT', API_BASE_URL . '/plugins/active');
// Add other endpoints as needed, e.g.:
// define('DEVICECLASSES_ENDPOINT', API_BASE_URL . '/deviceclasses');
//
//
?>
