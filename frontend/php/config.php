
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
// Network scanning / device import
define('SCAN_NETWORK_ENDPOINT', API_BASE_URL . '/scan/network');
define('SCAN_HOST_ENDPOINT', API_BASE_URL . '/scan/host');
define('SCAN_IMPORT_ENDPOINT', API_BASE_URL . '/scan/import');
define('SCAN_IMPORT_FILE_ENDPOINT', API_BASE_URL . '/scan/import-file');
define('SCAN_PROFILES_ENDPOINT', API_BASE_URL . '/scan/profiles');
define('SCAN_HOST_SSH_ENDPOINT', API_BASE_URL . '/scan/host-ssh');
define('SCAN_ANALYZE_ENDPOINT', API_BASE_URL . '/scan/analyze');
define('SCAN_EXECUTE_ENDPOINT', API_BASE_URL . '/scan/execute');
define('SCAN_CONNECTIONS_ENDPOINT', API_BASE_URL . '/scan/connections');
define('SCAN_CONNECTIONS_IMPORT_ENDPOINT', API_BASE_URL . '/scan/connections/import');
// Add other endpoints as needed, e.g.:
// define('DEVICECLASSES_ENDPOINT', API_BASE_URL . '/deviceclasses');
//
//
?>
