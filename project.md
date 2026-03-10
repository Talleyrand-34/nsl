Mapeo de redes, configuración declarativa de dispositivos de red. Actualmente tengo:

- Almacenamiento relacional de los datos de dispositivos de red

- Representación gráfica de estos elementos

    - Dispositivos de red  sus puertos

    - Conexiones

    - VLANs

    - Color en las conexiones dependiendo de la VLAN (en teoria puedes escoger si mostrar solo untagged o todas)     
Diagram



Las lineas de avance del proyecto son:

- Generación de configuraciones
    - Targets
        - OpenWRT
        - OPNSense
        - Fortinet
    - Ignorados
        - Cisco (No me gusta)
- Generación de los datos y metadatos iniciales con descubrimiento (Idea inicial pero me faltaba lo que desarrollé)
    - Creación de balizas de mapeo de red que tras su despliegue mapean los dispositivos de red y sus relaciones para alimentar a la herramienta
    - Esto puede evolucionar a balizas de testeo, donde cada una sabe donde está y la configuración de los dispositivos y puede lanzar una batería de tests para comprobar y auditar que el dispositivo está correctamente configurado.
    - Orquestración de tests, con erlang/elixir no solo comprobar la VLAN disponible si no que lanzar tests on coordinación entre nodos, por ejemplo comprobar que la CBAC está correctamente configurada
- Auditor de configuraciones
    - Un linux con openwrt, que pueda usar de forma maestra cada puerto y conectamos todos los puertos del target con el auditor y este manda paquetería y comprueba que el auditado devuelve el que debe y o devuelve el que no debe.
