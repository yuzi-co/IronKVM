const es = {
  translation: {
    feedback: {
      enabled: '{{name}} activado',
      disabled: '{{name}} desactivado',
      failed: 'La solicitud falló. Inténtalo de nuevo.',
      network:
        'No se pudo contactar con el dispositivo. Comprueba la conexión e inténtalo de nuevo.',
      saved: 'Guardado',
      timeout: 'El dispositivo tardó demasiado en responder. Inténtalo de nuevo.'
    },
    common: {
      copy: 'Copiar',
      copied: 'Copiado',
      copyFailed: 'No se pudo copiar. Selecciona el texto y cópialo a mano.',
      notUpdating: 'Sin actualizar: la última consulta falló.',
      off: 'Apagado',
      running: 'En ejecución',
      save: 'Guardar',
      cancel: 'Cancelar',
      delete: 'Eliminar',
      remove: 'Quitar'
    },
    head: {
      desktop: 'Escritorio remoto',
      login: 'Inicio de sesión',
      changePassword: 'Cambiar contraseña',
      terminal: 'Consola',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Contraseña cambiada. Inicia sesión con la nueva contraseña.',
      cookieRejected:
        'El navegador se ha negado a guardar la sesión. Una cookie que dejó una sesión HTTPS anterior no puede sustituirse por http sin cifrar. Borra las cookies de esta dirección, o abre una ventana privada, y vuelve a iniciar sesión.',
      login: 'Iniciar sesión',
      placeholderUsername: 'Introduce tu nombre de usuario',
      placeholderPassword: 'Introduce tu contraseña',
      placeholderCurrentPassword: 'Contraseña actual',
      placeholderPassword2: 'Introduce tu contraseña de nuevo',
      noEmptyUsername: 'El nombre de usuario no puede estar vacío',
      noEmptyPassword: 'La contraseña no puede estar vacía',
      passwordLength: 'La contraseña debe tener entre 8 y 72 caracteres',
      noAccount:
        'No se ha encontrado la cuenta. Por favor, recarga la página o recupera tu contraseña.',
      invalidUser: 'Nombre de usuario o contraseña incorrectos',
      locked: 'Demasiados inicios de sesión, inténtalo de nuevo más tarde',
      globalLocked: 'Sistema bajo protección, inténtelo nuevamente más tarde',
      error: 'Error inesperado',
      invalidCurrentPassword: 'La contraseña actual es incorrecta',
      changePassword: 'Cambiar contraseña',
      changePasswordDesc:
        'Para la seguridad de su dispositivo, por favor, modifique la contraseña de inicio de sesión en la web.',
      differentPassword: 'Las contraseñas no coinciden',
      illegalUsername: 'El  nombre de usuario contiene caracteres no permitidos',
      illegalPassword: 'La contraseña contiene caracteres no permitidos',
      forgetPassword: 'Contraseña olvidada',
      ok: 'Aceptar',
      cancel: 'Cancelar',
      loginButtonText: 'Iniciar sesión',
      tips: {
        reset1:
          'Para restablecer las contraseñas, mantén pulsado el botón BOOT del IronKVM durante 10 segundos.',
        reset3: 'Cuenta predeterminada de la interfaz web:',
        reset4: 'Cuenta predeterminada de SSH:',
        change1: 'Ten en cuenta que esta acción cambiará las siguientes contraseñas:',
        change2: 'Contraseña de acceso web',
        change3: 'Contraseña root del sistema (contraseña de acceso por SSH)',
        change4: 'Para restablecer las contraseñas, mantén pulsado el botón BOOT del IronKVM.',
        resetDocs: 'Para ver los pasos detallados, consulta la documentación del hardware:',
        hardwareDocs: 'Wiki de Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Configura el Wi-Fi para el IronKVM',
      success: 'Comprueba el estado de red del IronKVM y accede a la nueva dirección IP.',
      failed: 'La operación ha fallado, vuelve a intentarlo.',
      invalidMode:
        'El modo actual no admite la configuración de red. Vaya a su dispositivo y habilite el modo de configuración Wi-Fi.',
      confirmBtn: 'Aceptar',
      finishBtn: 'Finalizado',
      ap: {
        authTitle: 'Autenticación requerida',
        authDescription: 'Por favor ingrese la contraseña AP para continuar',
        authFailed: 'Contraseña AP no válida',
        passPlaceholder: 'AP contraseña',
        verifyBtn: 'Verificar'
      },
      ssidRequired: 'Introduzca el nombre de la red, hasta 32 caracteres',
      passwordLength:
        'La contraseña tiene de 8 a 63 caracteres. Déjela vacía para una red abierta.',
      passwordOptional: 'Contraseña (vacía para una red abierta)',
      lost: 'La placa dejó de responder. Puede que se haya unido a la red y cerrado su punto de acceso de configuración. Si el punto de acceso vuelve a aparecer, la conexión falló: vuelva a conectarse a él e inténtelo de nuevo.',
      done: 'Configuración terminada. Vuelva a conectar este dispositivo a su red habitual y abra la placa en su nueva dirección.'
    },
    screen: {
      codecNoWebrtcHevc: 'Este navegador no puede recibir H.265 por WebRTC',
      codecNoHevc: 'Este navegador no puede decodificar H.265',
      codecNote:
        'La placa tiene un solo codificador, así que esto cambia el flujo para todos los espectadores. Vuelva a conectar para aplicarlo a una sesión WebRTC en curso.',
      codec: 'Códec',
      updateFailed: 'No se aplicó el ajuste',
      scale: 'Escala',
      title: 'Pantalla',
      video: 'Modo de vídeo',
      videoDirectTips: 'Habilita HTTPS en "Ajustes > Dispositivo" para usar este modo',
      resolution: 'Resolución',
      ocr: {
        title: 'Leer texto (OCR)',
        tips: 'El texto se reconoce en este navegador. Puede corregirlo antes de copiarlo.',
        hint: 'Arrastre sobre el texto que desea leer. Pulse Esc para cancelar.',
        noPicture: 'Espere al vídeo y luego arrastre sobre el texto que desea leer.',
        cancel: 'Cancelar',
        language: 'Idioma',
        languages: {
          eng: 'Inglés'
        },
        preview: 'Área seleccionada',
        capturing: 'Capturando la pantalla...',
        loading: 'Cargando el reconocimiento de texto...',
        recognizing: 'Leyendo el texto...',
        noText: 'No se encontró texto en el área seleccionada.',
        copy: 'Copiar',
        copied: 'Copiado al portapapeles',
        copyFailed: 'No se pudo copiar al portapapeles',
        selectAgain: 'Seleccionar de nuevo',
        unsupported:
          'Este navegador no puede ejecutar el reconocimiento de texto. Necesita WebAssembly SIMD, que tienen los navegadores actuales.',
        captureFailed: 'No se pudo capturar la pantalla.',
        outside: 'El área seleccionada está fuera de la imagen.',
        recognizeFailed: 'Falló el reconocimiento de texto.'
      },
      controlRegion: {
        title: 'Calibración del ratón',
        description:
          'Utilice este ajuste cuando el dispositivo controlado use una resolución distinta de 16:9 y el cursor esté desalineado horizontal o verticalmente.',
        off: 'Desactivado',
        auto: 'Automático',
        autoWarning:
          'La calibración puede fallar si la aplicación del usuario tiene un fondo completamente negro.',
        manual: 'Manual',
        selectedResolution: 'Resolución del área seleccionada',
        unused: 'Sin usar',
        originalResolution: 'Resolución original',
        selectResolution: 'Seleccionar resolución original',
        addResolution: 'Añadir resolución personalizada',
        add: 'Añadir',
        duplicateResolution: 'Esta resolución ya existe.',
        width: 'Ancho',
        height: 'Alto',
        apply: 'Calcular y aplicar',
        invalidResolution: 'Introduzca una resolución original válida cuando el vídeo esté listo.',
        select: 'Seleccionar área',
        clear: 'Restaurar detección automática',
        saveFailed: 'No se pudo guardar el área de entrada.',
        tooSmall: 'El área seleccionada es demasiado pequeña.',
        previewUnavailable: 'Vista previa no disponible',
        clearConfirm: '¿Restaurar la detección automática de bordes negros?',
        dragHint: 'Arrastre para seleccionar el área del escritorio remoto',
        finish: 'Listo',
        confirm: 'Confirmar',
        cancel: 'Cancelar'
      },
      auto: 'Automático',
      autoTips:
        'En determinadas resoluciones pueden producirse rasgado de imagen (tearing) o desplazamiento del ratón. Prueba a ajustar la resolución del host remoto o desactiva el modo automático.',
      fps: 'FPS',
      customizeFps: 'Personalizar',
      quality: 'Calidad',
      qualityLossless: 'Máxima',
      qualityHigh: 'Alto',
      qualityMedium: 'Medio',
      qualityLow: 'Bajo',
      frameDetect: 'Detectar fotogramas',
      frameDetectTip:
        'Calcula la diferencia entre fotogramas. Para de transmitir vídeo cuando no se detectan cambios en la pantalla del host remoto.',
      resetHdmi: 'Reiniciar HDMI',
      mixedH264: {
        title: 'Conflicto de flujo H.264',
        description:
          'H.264 Direct y H.264 WebRTC se están utilizando al mismo tiempo. Esto puede causar tearing de pantalla o vídeo corrupto. Utilice solo un modo H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Error de conexión de WebRTC',
        description: 'Compruebe la conexión de red o cambie el modo de vídeo.'
      },
      captureStatus: {
        hdmiError: 'Error de imagen HDMI',
        unsupportedResolution: 'La resolución actual no es compatible',
        retrieving: 'Obteniendo pantalla...',
        changingResolution: 'Cambiando resolución...',
        updateFailed: 'La pantalla no puede actualizarse ahora',
        videoError: 'Error de visualización de video',
        noHdmi: 'No se detectó señal HDMI',
        unavailable: 'La pantalla no puede mostrarse ahora'
      },
      directConnectionFailed: 'Falló la conexión del flujo de vídeo'
    },
    keyboard: {
      close: 'Cerrar',
      title: 'Teclado',
      paste: 'Pegar',
      tips: 'Escribe el texto en el host como pulsaciones de teclas. Elige la distribución de teclado que usa el host.',
      placeholder: 'Por favor, introduce el texto',
      submit: 'Enviar',
      virtual: 'Teclado virtual',
      readClipboard: 'Leer del portapapeles',
      clipboardPermissionDenied:
        'Permiso de portapapeles denegado. Por favor, permite el acceso al portapapeles en tu navegador.',
      clipboardReadError: 'Error al leer del portapapeles',
      mediaKeys: {
        title: 'Teclas multimedia',
        mute: 'Silenciar',
        volumeDown: 'Bajar volumen',
        volumeUp: 'Subir volumen',
        previous: 'Pista anterior',
        playPause: 'Reproducir o pausar',
        next: 'Pista siguiente',
        stop: 'Detener'
      },
      pasting: {
        layout: 'Distribución de teclado del host',
        layouts: {
          us: 'Inglés (EE. UU.)',
          uk: 'Inglés (Reino Unido)',
          de: 'Alemán',
          fr: 'Francés',
          es: 'Español',
          it: 'Italiano',
          ptBr: 'Portugués (Brasil)',
          se: 'Sueco / finés',
          ru: 'Ruso',
          ja: 'Japonés',
          ko: 'Coreano'
        },
        speed: 'Velocidad de escritura',
        speeds: {
          fast: 'Rápida',
          normal: 'Normal',
          slow: 'Lenta'
        },
        estimate: 'Tiempo de escritura: unos {{duration}}',
        untypeable: 'Caracteres que esta distribución no puede escribir: {{count}}',
        untypeableAt: 'línea {{line}}, columna {{column}}',
        skipUntypeable: 'Escribir el resto',
        shortcut: '{{shortcut}} escribe el portapapeles en el host directamente.',
        clipboardUnavailable:
          'El navegador solo permite que una página lea el portapapeles por HTTPS. Pega el texto en el cuadro con Ctrl+V.',
        clipboardEmpty: 'El portapapeles no contiene texto.',
        tooLong: 'El texto es demasiado largo. El límite es de {{max}} caracteres.',
        inProgress: 'Ya se está escribiendo un texto pegado.',
        typing: 'Escribiendo en el host',
        done: 'Texto escrito',
        canceled: 'Pegado cancelado',
        failed: 'El pegado ha fallado',
        cancel: 'Cancelar',
        controlBusy: 'Otro controlador está usando el teclado.',
        hidError: 'No se pudieron enviar las pulsaciones al host.'
      },
      shortcut: {
        sendFailed: 'No enviado: la conexión de entrada no está disponible',
        title: 'Atajos',
        custom: 'Personalizado',
        capture: 'Haga clic aquí para capturar el acceso directo',
        clear: 'Borrar',
        save: 'Guardar',
        captureTips:
          'Capturar teclas del sistema (como la tecla Windows) requiere permiso de pantalla completa.',
        enterFullScreen: 'Alternar el modo de pantalla completa.'
      },
      leaderKey: {
        saveFailed: 'No se pudo guardar la tecla líder',
        title: 'Tecla líder',
        desc: 'Omite las restricciones del navegador y envía accesos directos al sistema directamente al host remoto.',
        howToUse: 'Cómo utilizar',
        simultaneous: {
          title: 'Modo Simultáneo',
          desc1: 'Mantenga pulsada la tecla líder y luego pulse el atajo.',
          desc2: 'Intuitivo, pero puede entrar en conflicto con los atajos del sistema.'
        },
        sequential: {
          title: 'Modo Secuencial',
          desc1:
            'Pulse la tecla líder → pulse el atajo en secuencia → vuelva a pulsar la tecla líder.',
          desc2: 'Requiere más pasos, pero evita por completo conflictos del sistema.'
        },
        enable: 'Habilitar tecla líder',
        tip: 'Al asignarse como tecla líder, esta tecla funciona únicamente como activador de atajos y pierde su comportamiento predeterminado.',
        placeholder: 'Pulse la tecla líder',
        shiftRight: 'Shift derecho',
        ctrlRight: 'Ctrl derecho',
        metaRight: 'Win derecho',
        submit: 'Enviar',
        recorder: {
          rec: 'REC',
          activate: 'Activar teclas',
          input: 'Por favor presione el atajo...'
        }
      }
    },
    mouse: {
      title: 'Ratón',
      cursor: 'Estilo de cursor',
      default: 'Cursor por defecto',
      pointer: 'Cursor de puntero',
      cell: 'Cursor de celda',
      text: 'Cursor de texto',
      grab: 'Cursor de agarre',
      hide: 'Ocultar cursor',
      mode: 'Modo de ratón',
      absolute: 'Modo absoluto',
      relative: 'Modo relativo',
      absoluteShort: 'Absoluto',
      relativeShort: 'Relativo',
      touch: 'Modo táctil',
      touchShort: 'Táctil',
      absoluteStalled: 'El host está ignorando el ratón absoluto',
      absoluteStalledDesc:
        'El host ha dejado de recoger los informes del ratón absoluto, así que los movimientos del puntero se pierden. El teclado no se ve afectado. Restablecer el USB suele solucionarlo; el modo relativo usa otro endpoint.',
      useRelative: 'Cambiar a modo relativo',
      direction: 'Dirección de la rueda de desplazamiento',
      scrollUp: 'Igual que este equipo',
      scrollDown: 'Invertido (desplazamiento natural)',
      speed: 'Velocidad de la rueda de desplazamiento',
      fast: 'Rápida',
      slow: 'Lenta',
      requestPointer:
        'Usando modo relativo. Por favor, haz clic en el escritorio para obtener el cursor del ratón.',
      resetHid: 'Restablecer HID',
      hidOnly: {
        switchFailed: 'No se pudo cambiar el modo. Compruebe la conexión e inténtelo de nuevo.',
        title: 'Modo solo HID',
        desc: 'Si tu ratón y teclado dejan de responder y restablecer el HID no ayuda, podría ser un problema de compatibilidad entre el IronKVM y el dispositivo. Prueba a habilitar el modo sólo HID para mejorar la compatibilidad.',
        tip1: 'Habilitar el modo sólo HID desmontará el disco virtual y la red virtual',
        tip2: 'En modo sólo HID, el montaje de imágenes está deshabilitado',
        rebuild: 'Cambiar de modo reconstruye la conexión USB. El IronKVM no se reinicia',
        enable: 'Habilitar modo sólo HID',
        disable: 'Desactivar modo sólo HID'
      },
      resetHidDone: 'HID USB reiniciado',
      resetHidFailed: 'No se pudo reiniciar el HID USB'
    },
    image: {
      delete: 'Eliminar',
      inUse: 'En uso. Expúlsela antes de eliminarla.',
      retry: 'Reintentar',
      loadFailed: 'No se pudo cargar la lista de imágenes',
      readOnlyLocked: 'Expulse el disco para cambiarlo. Se aplica al insertar una imagen.',
      title: 'Imágenes',
      loading: 'Cargando...',
      empty: 'No se ha encontrado nada',
      mountMode: 'Modo de montaje',
      mountFailed: 'Fallo al montar',
      mountDesc:
        'En algunos sistemas, es necesario expulsar el disco virtual del host remoto antes de montar una imagen.',
      unmountFailed: 'Fallo al desmontar',
      unmountDesc:
        'En algunos sistemas, es necesario expulsar manualmente el disco virtual desde el host remoto antes de desmontar la imagen.',
      refresh: 'Actualizar la lista de imágenes',
      disk: 'Disco',
      cdrom: 'CD',
      driveEmpty: 'Vacío',
      eject: 'Expulsar',
      readOnly: 'Solo lectura',
      readOnlyTip: 'Se aplica a la próxima imagen que se inserte en el disco.',
      noDrives: 'No hay unidades virtuales. Activa el disco virtual en Ajustes.',
      insertFailed: 'Error al insertar',
      ejectFailed: 'Error al expulsar',
      insertInto: 'Insertar en {{drive}}. Haz clic para cambiar.',
      loadedIn: 'En la unidad {{drive}}',
      attention: 'Atención',
      deleteConfirm: '¿Estás seguro de que deseas eliminar esta imagen?',
      okBtn: 'Sí',
      cancelBtn: 'No',
      deleteFailed: 'Error al eliminar',
      ventoy: {
        statusNoKernel: 'No compatible con este firmware',
        statusNotInstalled: 'No instalado',
        statusReady: 'Listo',
        statusSelected: 'Imágenes seleccionadas: {{count}}',
        statusInDrive: 'En la unidad de disco, {{size}}',
        noKernel:
          'El kernel de este firmware no admite device-mapper, así que Ventoy no se puede usar hasta instalar una imagen que lo admita.',
        installDesc: 'Arranca el host desde varias imágenes en un solo disco, sin copiarlas.',
        install: 'Instalar',
        installing: 'Descargando Ventoy, unos 20 MB. Puede tardar unos minutos.',
        needsData: 'Ventoy necesita una imagen de IronKVM con la partición /data montada.',
        uninstall: 'Desinstalar',
        uninstallConfirm: '¿Eliminar los archivos de Ventoy?',
        noImages: 'No hay imágenes para el disco Ventoy.',
        onDisk: 'En el disco Ventoy',
        missing: 'Falta: {{file}}',
        remove: 'Quitar del disco Ventoy',
        setHint:
          'El conjunto de imágenes solo cambia mientras el disco Ventoy no está en ninguna unidad.',
        useAsDisk: 'Usar como disco virtual',
        failed: 'Falló la solicitud de Ventoy',
        secureBoot:
          'Con Secure Boot activado, el host debe registrar una vez la clave de Ventoy en MokManager. El archivo de clave ENROLL_THIS_KEY_IN_MOKMANAGER.cer está en la partición VTOYEFI.',
        readOnly:
          'El host ve el disco como de solo lectura, así que la persistencia de Ventoy y ventoy.json en la unidad no funcionan.'
      },
      tips: {
        title: 'Cómo subir imágenes',
        usb1: 'Conecta el IronKVM a tu computadora mediante USB.',
        usb2: 'Asegúrate de que el disco virtual esté montado (Ajustes - Disco Virtual).',
        usb3: 'Abre el disco virtual en tu computadora y copia el archivo de imagen en el directorio raíz del disco virtual.',
        scp1: 'Asegúrate de que el IronKVM y tu computadora estén en la misma red local.',
        scp2: 'Abre una terminal en tu computadora y usa el comando SCP para subir el archivo de imagen al directorio /data en el IronKVM.',
        scp3: 'Ejemplo: scp tu-ruta-de-imagen root@tu-ip-del-nanokvm:/data',
        tfCard: 'Tarjeta SD',
        tf1: 'Este método es compatible con el sistema Linux',
        tf2: 'Obtén la tarjeta SD del IronKVM (para la versión FULL, desmonta la carcasa primero).',
        tf3: 'Inserta la tarjeta SD en un lector de tarjetas y conéctalo a tu computadora.',
        tf4: 'Copia el archivo de imagen en el directorio /data de la tarjeta SD.',
        tf5: 'Inserta la tarjeta SD en el IronKVM.'
      }
    },
    script: {
      title: 'Script',
      upload: 'Subir',
      run: 'Ejecutar',
      runBackground: 'Ejecutar en segundo plano',
      runFailed: 'Ejecución fallida',
      attention: 'Atención',
      delDesc: '¿Estás seguro de que deseas eliminar este archivo?',
      confirm: 'Sí',
      cancel: 'No',
      delete: 'Eliminar',
      close: 'Cerrar',
      empty: 'Aún no hay scripts. Suba un archivo .sh o .py para ejecutarlo en la placa.',
      loadFailed: 'No se pudieron cargar los scripts',
      uploaded: 'Script subido',
      uploadFailed: 'No se pudo subir el script',
      started: 'Script iniciado en segundo plano',
      deleteFailed: 'No se pudo eliminar el script',
      waitLimit: 'Esperando a que termine el script, hasta {{minutes}} minutos.',
      timedOut:
        'El script tardó más de {{minutes}} minutos y esta página dejó de esperar. Puede que siga ejecutándose en la placa.'
    },
    terminal: {
      invalidBaud: 'Esta velocidad en baudios no es compatible.',
      invalidPort: 'Introduzca una ruta de dispositivo en /dev, como /dev/ttyS1.',
      invalidSettings: 'Ajustes del puerto serie no válidos. Esta es la consola de la placa.',
      disconnected: 'Desconectado. Pulse Intro para volver a conectar.',
      title: 'Consola',
      nanokvm: 'Consola del IronKVM',
      serial: 'Consola del Puerto Serie',
      serialPort: 'Puerto Serie',
      serialPortPlaceholder: 'Por favor, introduce el puerto serie',
      baudrate: 'Tasa de baudios',
      parity: 'Paridad',
      parityNone: 'Ninguna',
      parityEven: 'Par',
      parityOdd: 'Impar',
      flowControl: 'Control de flujo',
      flowControlNone: 'Ninguno',
      flowControlSoft: 'Software',
      flowControlHard: 'Hardware',
      dataBits: 'Bits de datos',
      stopBits: 'Bits de parada',
      confirm: 'Confirmar'
    },
    wol: {
      no: 'No',
      yes: 'Sí',
      deleteConfirm: '¿Eliminar esta dirección guardada?',
      delete: 'Eliminar',
      wake: 'Despertar',
      rename: 'Cambiar nombre',
      showMac: 'Mostrar dirección MAC',
      showName: 'Mostrar nombre',
      requestFailed: 'No se pudo contactar con el dispositivo para enviar la orden',
      deleteFailed: 'No se pudo eliminar',
      renameFailed: 'No se pudo cambiar el nombre',
      title: 'Wake-on-LAN',
      sending: 'Enviando comando...',
      sent: 'Comando enviado',
      input: 'Por favor, introduce la dirección MAC',
      ok: 'Aceptar'
    },
    download: {
      uploadFailed: 'La subida falló',
      uploadSuccess: 'Subida completada',
      uploading: 'Subiendo: {{file}}',
      downloadingPercent: 'Descargando ({{percent}}): {{file}}',
      downloading: 'Descargando: {{file}}',
      title: 'Descargador de imágenes',
      input: 'Por favor, introduce la URL de una imagen remota',
      ok: 'Aceptar',
      disabled: 'La partición /data es de sólo lectura, no se puede descargar la imagen',
      uploadbox: 'Suelte el archivo aquí o haga clic para seleccionar',
      inputfile: 'Por favor ingrese el archivo de imagen',
      NoISO: 'Sin ISO',
      sha256: 'SHA-256 (opcional)',
      sha256Placeholder: 'Introduzca una suma de comprobación SHA-256 de 64 caracteres',
      invalidSHA256: 'SHA-256 debe ser una cadena hexadecimal de 64 caracteres',
      failed: 'Descarga fallida',
      success: 'Descarga correcta',
      checksumFailed: 'Descarga fallida: error en la verificación SHA-256',
      cancel: 'Cancelar',
      cancelFailed: 'No se pudo cancelar la descarga',
      bootMenu: 'Menú de arranque (netboot.xyz)',
      bootMenuPresent: '{{file}} ya está en el dispositivo, con la suma correcta',
      bootMenuDesc: 'Descargar la ISO de netboot.xyz, con la suma comprobada, para el CD virtual'
    },
    power: {
      title: 'Encender / Apagar',
      showConfirm: 'Confirmación',
      showConfirmTip:
        'Preguntar antes de una pulsación corta. El reinicio y la pulsación larga siempre preguntan.',
      reset: 'Reiniciar',
      power: 'Encender / Apagar',
      powerShort: 'Encender / Apagar (pulsación corta)',
      powerLong: 'Encendido/Apagado (pulsación larga)',
      resetConfirm: '¿Desea proceder con la operación de reinicio?',
      powerConfirm: '¿Desea proceder con la operación de encendido?',
      okBtn: 'Sí',
      cancelBtn: 'No',
      hostOs: 'SO del host',
      hostOsTip: 'Se envían como teclas USB. El host decide qué hacen.',
      sleep: 'Suspender',
      wake: 'Despertar',
      wakeKey: 'Despertar con Mayús',
      powerDown: 'Apagar',
      sleepConfirm: '¿Suspender el host?',
      powerDownConfirm: '¿Enviar la tecla de apagado al host?',
      wakeTip:
        'Un host suspendido suele ignorar Despertar del dispositivo que lo suspendió. Despertar con Mayús pulsa una tecla del teclado, que más hosts aceptan.',
      led: 'LED de encendido',
      ledOn: 'Encendido',
      ledOff: 'Apagado',
      ledUnknown: 'Desconocido',
      ledConnected: 'LED de encendido conectado',
      ledConnectedTip:
        'Actívalo solo si el conector del LED de encendido del host está cableado a la placa. Sin él, el estado de encendido es desconocido.',
      ledConnectedFailed: 'No se pudo guardar el ajuste del LED de encendido',
      powerLongConfirm:
        '¿Mantener pulsado el botón de encendido {{seconds}} s? Corta la alimentación sin apagar el sistema.',
      done: 'Botón pulsado',
      failed: 'No se pudo pulsar el botón'
    },
    settings: {
      title: 'Ajustes',
      nav: {
        system: 'Sistema',
        network: 'Red',
        access: 'Acceso',
        integrations: 'Integraciones',
        boot: 'Arranque',
        browser: 'Este navegador',
        search: 'Buscar un ajuste',
        noMatch: 'Ningún ajuste coincide',
        locked:
          'Hay una operación en curso. Las demás páginas y el cierre no están disponibles hasta que termine.',
        vpnProvider: 'Proveedor de VPN'
      },
      mcp: {
        keyNote:
          'MCP usa su propia clave API, mostrada abajo. Las claves de la página Claves API no funcionan aquí.',
        title: 'Servicio MCP',
        service: 'Control remoto MCP',
        serviceDesc:
          'Permitir que clientes MCP de confianza controlen el teclado y el ratón y realicen capturas de pantalla',
        securityWarning:
          'Cualquier persona con esta clave API puede controlar el host remoto y ver su pantalla. Utiliza HTTPS y actívalo solo en redes de confianza.',
        endpoint: 'Punto de conexión',
        apiKey: 'Clave API',
        regenerateConfirmTitle: '¿Volver a generar la clave API de MCP?',
        regenerateConfirmDesc: 'La clave actual dejará de funcionar inmediatamente.',
        enableConfirmTitle: '¿Activar el control MCP externo?',
        enableConfirmDesc:
          'Al activar MCP se detendrá PicoClaw y se cerrará cualquier sesión activa de PicoClaw.',
        failed: 'Error en la operación MCP',
        copyFailed: 'Error al copiar. Copia manualmente.',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar',
        showKey: 'Mostrar clave',
        hideKey: 'Ocultar clave',
        regenerateKey: 'Generar clave nueva'
      },
      redfish: {
        example: 'Ejemplo',
        title: 'Redfish',
        service: 'Servicio Redfish',
        serviceDesc:
          'La API Redfish de DMTF, para control de encendido, medios virtuales y estado desde herramientas como redfishtool y Ansible. Al desactivarla se cierran todas las sesiones Redfish.',
        endpoint: 'Raíz del servicio',
        httpsOn: 'La placa sirve HTTPS, que la mayoría de herramientas Redfish necesitan.',
        httpsOff:
          'La placa sirve HTTP sin cifrar. La mayoría de herramientas Redfish necesitan HTTPS: actívalo en "Ajustes > Red".',
        credentials:
          'Redfish acepta las cuentas del KVM, con autenticación Basic o una sesión Redfish, y claves API enviadas como X-Auth-Token. Las claves API se gestionan en la página Claves API.',
        powerActions: 'Acciones de encendido',
        powerActionsDesc:
          'Los tipos de reinicio que se ofrecen ahora. On, ForceOff y GracefulShutdown necesitan el estado de encendido, así que solo se ofrecen cuando "LED de encendido conectado" está activado en el menú de encendido.',
        sessions: 'Sesiones',
        noSessions: 'No hay sesiones Redfish abiertas',
        created: 'Creada',
        lastUsed: 'Último uso',
        refresh: 'Actualizar',
        end: 'Cerrar',
        endConfirmTitle: '¿Cerrar esta sesión Redfish?',
        endConfirmDesc:
          'Su token deja de funcionar al instante. El cliente tendrá que volver a iniciar sesión.',
        failed: 'Error en la operación Redfish',
        copyFailed: 'Error al copiar. Copia manualmente.',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar'
      },
      ipmi: {
        copyBeforeSave:
          'Copia la contraseña ahora. Una vez guardada, no se puede volver a mostrar.',
        noLogin:
          'IPMI está activado, pero ninguna cuenta activa tiene contraseña IPMI, así que nadie puede iniciar sesión. Define una abajo.',
        title: 'IPMI',
        warning:
          'La autenticación IPMI es débil por diseño. Cualquiera que llegue a la placa y conozca un nombre de usuario puede obtener un hash de la contraseña IPMI de ese usuario e intentar descifrarlo sin conexión. Use contraseñas generadas, active IPMI solo en una red de confianza y prefiera Redfish sobre HTTPS cuando la herramienta lo admita.',
        service: 'IPMI sobre LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) en el puerto UDP 623, para la alimentación y el estado del host. IPMI 1.5 y el conjunto de cifrado 0 se rechazan. Desactivarlo cierra todas las sesiones IPMI.',
        example: 'Ejemplo',
        copyFailed: 'Error al copiar. Copie manualmente.',
        ledOn: 'Están disponibles el estado, on, off, soft, cycle y reset.',
        ledOff:
          '"LED de encendido conectado" está desactivado en el menú de encendido, así que el estado de alimentación es desconocido. Solo funciona "power reset": status, on, off, soft y cycle se rechazan.',
        accounts: 'Cuentas',
        accountsDesc:
          'IPMI inicia sesión con las cuentas del KVM, cada una con su propia contraseña IPMI, distinta de la contraseña web. Los administradores obtienen ADMINISTRATOR. Los usuarios obtienen USER: pueden leer el estado de alimentación con "-L USER", pero no cambiarlo.',
        passwordSet: 'Contraseña IPMI definida',
        passwordNotSet: 'Sin contraseña IPMI: no puede iniciar sesión por IPMI',
        nameTooLong: 'El nombre tiene más de 16 caracteres, lo que IPMI no permite',
        accountDisabled: 'La cuenta está desactivada',
        setPassword: 'Definir contraseña',
        changePassword: 'Cambiar contraseña',
        remove: 'Quitar',
        removeConfirmTitle: '¿Quitar la contraseña IPMI de {{user}}?',
        removeConfirmDesc:
          'La cuenta ya no podrá iniciar sesión por IPMI y sus sesiones IPMI terminarán.',
        passwordTitle: 'Contraseña IPMI de {{user}}',
        passwordDesc:
          'De 12 a 20 caracteres ASCII imprimibles, distinta de la contraseña web. IPMI necesita que la placa guarde la contraseña de forma que pueda volver a leerla, así que use una que no se use en ningún otro sitio. Cópiela antes de guardar: no se volverá a mostrar.',
        passwordPlaceholder: 'Contraseña IPMI',
        generate: 'Generar',
        copy: 'Copiar',
        save: 'Guardar',
        passwordLength: 'Use de 12 a 20 caracteres.',
        passwordChars: 'Use solo caracteres ASCII imprimibles.',
        saved: 'Contraseña IPMI guardada',
        failed: 'La operación IPMI falló',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar'
      },
      ssh: {
        service: 'Servidor SSH',
        serviceDesc: 'Iniciar sshd ahora y en cada arranque',
        failed: 'No se pudo cargar la configuración de SSH',
        rootDefault: 'root aún tiene la contraseña de fábrica',
        rootEmpty: 'root no tiene contraseña',
        rootWarning:
          'Cualquiera que llegue a la consola o a SSH puede entrar como root. Defina una contraseña en {{account}} > {{password}}: para el propietario del dispositivo también cambia la de root.',
        connection: 'Conexión',
        command: 'Entrar como root',
        port: 'Puerto',
        viaVpn: 'Por {{name}}',
        notRunning: 'sshd no está en marcha. Active el servidor SSH para conectarse.',
        hostKeys: 'Huellas de las claves de host',
        hostKeysDesc: 'Compárelas con lo que muestra ssh en la primera conexión.',
        noHostKeys: 'Aún no hay claves de host. sshd las crea la primera vez que arranca.',
        keys: 'Claves autorizadas',
        keysDesc:
          'Claves públicas que pueden entrar como root. Se guardan en la partición de datos, así que las actualizaciones las conservan.',
        noKeys: 'Aún no hay claves autorizadas.',
        noComment: 'sin comentario',
        addPlaceholder:
          'Pegue una clave pública, por ejemplo el contenido de ~/.ssh/id_ed25519.pub',
        add: 'Añadir clave',
        added: 'Clave añadida',
        removed: 'Clave eliminada',
        deleteConfirm: '¿Eliminar esta clave?',
        deleteConfirmDesc: 'Ya no podrá entrar. Las sesiones abiertas siguen abiertas.',
        invalidKey: 'Esto no es una clave pública. Pegue una sola línea de un archivo .pub.',
        keyOptions: 'Aquí no se aceptan claves con opciones como command= o from=.',
        duplicateKey: 'Esta clave ya está autorizada.',
        lastKey:
          'La última clave no se puede eliminar mientras el acceso solo con claves esté activado.',
        keysOnly: 'Solo claves',
        keysOnlyDesc:
          'Desactivar el acceso con contraseña y keyboard-interactive. Las sesiones abiertas siguen abiertas.',
        keysOnlyNeedsKey: 'Añada primero una clave autorizada, o nadie podría entrar.',
        keysOnlyOn: 'Acceso con contraseña desactivado',
        keysOnlyOff: 'Acceso con contraseña activado',
        notHonoured:
          'El sshd de esta imagen no lee este ajuste, así que el acceso con contraseña sigue activo.',
        reloadFailed:
          'Guardado, pero no se pudo recargar sshd. Se aplicará la próxima vez que arranque sshd.',
        notApplied:
          'sshd aún acepta contraseñas. Apague y encienda el servidor SSH para aplicar el ajuste.'
      },
      vnc: {
        address: 'Dirección',
        certHint:
          'VeNCrypt X509Plain usa el certificado autofirmado del dispositivo, así que el cliente avisa en la primera conexión. Acéptalo, o guarda el certificado desde la dirección HTTPS de esta página y pásalo a TigerVNC con -X509CA=<archivo>.',
        example: 'Ejemplo',
        title: 'VNC',
        service: 'Servidor VNC',
        serviceDesc:
          'Permite que un cliente VNC, como TigerVNC o Remmina, vea y controle el host. El cliente debe admitir la codificación Tight. Una sesión a la vez.',
        credentials:
          'Inicie sesión con una cuenta KVM. La conexión se cifra con el certificado TLS de la placa (VeNCrypt X509Plain).',
        port: 'Puerto',
        portDesc: 'El puerto TCP en el que escucha el servidor.',
        maxFps: 'Límite de fotogramas',
        maxFpsDesc: 'El máximo de fotogramas por segundo que se envían a un cliente.',
        vncAuth: 'Autenticación VNC simple',
        vncAuthDesc:
          'Para clientes sin VeNCrypt. Comprueba una contraseña VNC aparte en lugar de una cuenta.',
        vncAuthWarning:
          'La autenticación VNC simple no cifra la conexión. Cualquiera en la ruta de red puede ver la pantalla y las pulsaciones de teclas. Úsela solo en una red de confianza.',
        password: 'Contraseña VNC',
        passwordSet: 'Hay una contraseña establecida. Escriba una nueva para cambiarla.',
        passwordInvalid: 'La contraseña VNC debe tener de 6 a 8 caracteres.',
        save: 'Guardar',
        saved: 'Ajustes guardados',
        state: 'Estado',
        listening: 'Escuchando en el puerto {{port}}',
        notListening: 'No escucha',
        noSession: 'No hay ninguna sesión abierta',
        client: 'Cliente',
        user: 'Usuario',
        method: 'Autenticación',
        methodVencrypt: 'Cuenta sobre TLS',
        methodVnc: 'Contraseña VNC',
        since: 'Conectado desde',
        resolution: 'Resolución',
        framesSent: 'Fotogramas enviados',
        lastError: 'La última sesión terminó: {{error}}',
        refresh: 'Actualizar',
        disconnect: 'Desconectar',
        disconnectConfirmTitle: '¿Terminar la sesión VNC?',
        disconnectConfirmDesc:
          'El cliente se desconecta de inmediato y se sueltan todas las teclas y botones que mantenga pulsados.',
        failed: 'La operación VNC falló',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Watchdog del host',
        serviceDesc:
          'Si el host debería estar encendido y su imagen no cambia, o no hay señal HDMI, durante el tiempo de espera, la placa pulsa reset o apaga y enciende el host.',
        stillWarning:
          'Un host cuya pantalla entra en reposo, o cuya imagen permanece fija mientras trabaja, parece colgado. Desactive el reposo de pantalla en el host o indique una dirección de ping.',
        ledHint:
          '"LED de encendido conectado" está desactivado en el menú de energía. El watchdog no ve cuándo el host está apagado, así que lo trata como siempre encendido.',
        timeout: 'Tiempo de espera',
        timeoutDesc:
          'Cuánto tiempo puede el host no dar señales de vida antes de que el watchdog actúe.',
        action: 'Acción',
        actionDesc:
          'El ciclo de energía mantiene pulsado el botón de encendido 5 segundos y luego lo pulsa de nuevo.',
        actionReset: 'Reiniciar',
        actionPower: 'Ciclo de energía',
        cooldown: 'Tiempo de reposo',
        cooldownDesc: 'El tiempo mínimo entre dos acciones.',
        maxPerHour: 'Acciones por hora',
        maxPerHourDesc: 'El máximo de acciones en una hora.',
        pingHost: 'Dirección de ping',
        pingHostDesc:
          'La dirección IP del host. Una respuesta cuenta como señal de vida. Déjela vacía para no hacer ping.',
        pingHostInvalid: 'Introduzca una dirección IPv4 o IPv6.',
        minutes: 'min',
        save: 'Guardar',
        saved: 'Guardado',
        state: 'Detector',
        status: {
          off: 'Desactivado',
          watching: 'Vigilando',
          hostOff: 'Host apagado',
          captureOff: 'Captura HDMI desactivada',
          cooldown: 'En reposo',
          capped: 'Límite por hora alcanzado',
          acting: 'Actuando'
        },
        signal: 'Señal HDMI',
        yes: 'Sí',
        no: 'No',
        led: 'LED de encendido',
        on: 'Encendido',
        off: 'Apagado',
        ledNotConnected: 'No conectado',
        ping: 'Ping',
        pingNotSet: 'Sin definir',
        pingReply: 'Responde',
        pingNoReply: 'Sin respuesta',
        lastChange: 'Último cambio de imagen',
        never: 'Nunca',
        actsIn: 'Actúa en',
        actionsLastHour: 'Acciones en la última hora',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Registro',
        noLog: 'El watchdog aún no ha actuado.',
        refresh: 'Actualizar',
        reasonFrozen: 'La imagen no cambió',
        reasonNoSignal: 'Sin señal HDMI',
        stuckFor: 'sin señales de vida durante {{duration}}',
        pressFailed: 'La pulsación falló: {{error}}',
        noScreenshot: 'Sin captura',
        failed: 'Error en la operación del watchdog',
        powerNeedsLed:
          'El ciclo de encendido necesita «LED de encendido conectado» en el menú de energía.',
        noLedConfirmTitle: '¿Activar el watchdog sin el LED de encendido?',
        noLedConfirmDesc:
          'La placa no puede ver cuándo el host está apagado, así que lo trata como siempre encendido. Si apaga el host, el watchdog pulsa reinicio al pasar el tiempo de espera. Conecte el LED de encendido para evitarlo.',
        noLedConfirmOk: 'Activar',
        cancel: 'Cancelar'
      },
      netboot: {
        title: 'Arranque por red',
        description:
          'Arrancar el host desde la red: iPXE y un menú de las imágenes del KVM por el enlace de red USB, o netboot.xyz por proxy DHCP en la LAN.',
        addon: 'dnsmasq y archivos de arranque',
        addonDesc:
          'Instalados en /data: dnsmasq desde Alpine, iPXE y netboot.xyz desde sus versiones publicadas, cada uno comprobado con su suma de verificación.',
        install: 'Instalar',
        installing: 'Instalando. Puede tardar unos minutos.',
        uninstall: 'Desinstalar',
        uninstallConfirm:
          '¿Desactivar el arranque por red y eliminar dnsmasq y los archivos de arranque?',
        needsData:
          'El arranque por red necesita una imagen de IronKVM con la partición /data montada.',
        usb: 'En el enlace de red USB',
        usbDesc:
          'Mientras el enlace de red USB está activo, dnsmasq lo atiende en lugar de udhcpd. El host recibe su única dirección sin router ni servidor DNS, iPXE para su arquitectura y un menú de las imágenes ISO del KVM.',
        linkOff: 'El enlace de red USB está desactivado. Actívalo en Dispositivo, Red USB.',
        menuUrl: 'Menú',
        leases: 'Concesión del host',
        noLeases: 'Ninguna todavía',
        netbootxyzNote:
          'netboot.xyz en el menú se carga desde internet, al que el enlace USB no llega. El host necesita internet en otro puerto de red para usarlo.',
        lan: 'Proxy DHCP en la LAN',
        lanDesc:
          'Responde a los clientes PXE de la LAN con netboot.xyz, que luego carga su menú desde internet. Nunca asigna direcciones y no sirve las imágenes del KVM.',
        lanWarning:
          'Se ofrece netboot.xyz a todos los clientes PXE de esta LAN, no solo al host. Actívalo solo en una red que controles.',
        lanConfirm: '¿Activar el proxy DHCP en la LAN?',
        lanInterface: 'LAN',
        running: 'En ejecución',
        stopped: 'Detenido',
        images: 'Imágenes en el menú',
        noImages: 'No hay imágenes ISO en el directorio de imágenes.',
        boots: 'Arranques recientes',
        noBoots: 'El host todavía no ha descargado nada.',
        log: 'Registro de dnsmasq',
        refresh: 'Actualizar',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar',
        failed: 'Error en la operación de arranque por red'
      },
      about: {
        title: 'Sobre IronKVM',
        information: 'Información',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versión de la aplicación',
        applicationTip: 'Versión de la aplicación web IronKVM',
        image: 'Versión de la imagen',
        imageTip:
          'Imagen de tarjeta de IronKVM y la imagen de sistema de NanoKVM en la que se basa',
        kernel: 'Versión del kernel',
        kernelTip: 'Versión del kernel de Linux en ejecución',
        deviceKey: 'Clave del dispositivo',
        videoMemory: 'Memoria de vídeo',
        videoMemoryTip:
          'Memoria reservada para la captura de vídeo. No se comparte con el resto del sistema.',
        videoMemoryGenerations_one:
          '{{count}} sesión anterior de IronKVM está reteniendo memoria de vídeo',
        videoMemoryGenerations_other:
          '{{count}} sesiones anteriores de IronKVM están reteniendo memoria de vídeo',
        videoMemoryReboot: 'Reinicia para recuperarla.',
        community: 'Comunidad',
        hostname: 'Nombre del host',
        hostnameUpdated: 'Nombre del host actualizado. Reinicia para aplicar.',
        ipType: {
          Wired: 'Cableada',
          Wireless: 'Inalámbrica',
          Other: 'Otra'
        },
        hostnameInvalid:
          'Use letras, dígitos y guiones, hasta 63 por parte separada por puntos. Sin guion al principio ni al final de una parte.',
        hostnameFailed: 'No se pudo cambiar el nombre de host',
        editHostname: 'Editar nombre de host',
        docs: 'Documentación',
        hardware: 'Hardware',
        hardwareFaq: 'Preguntas frecuentes del hardware',
        disclaimer:
          'IronKVM: firmware comunitario reforzado para el Sipeed NanoKVM. Sin relación con Sipeed.',
        basedOn: 'basado en NanoKVM {{version}}'
      },
      preferences: {
        title: 'Preferencias'
      },
      performance: {
        title: 'Rendimiento'
      },
      appearance: {
        thisBrowser: 'Este navegador',
        thisBrowserDesc:
          'Se guarda solo en este navegador. Los demás navegadores conservan la suya.',
        deviceWide: 'Dispositivo',
        deviceWideDesc: 'Se guarda en el dispositivo. Se aplica a todos los que lo abren.',
        language: 'Idioma',
        languageDesc: 'Seleccionar el idioma de la interfaz',
        webTitle: 'Título web',
        webTitleDesc: 'Personaliza el título de la página web',
        menuBar: {
          title: 'Barra de menú',
          mode: 'Modo de visualización',
          modeDesc: 'Mostrar barra de menú en la pantalla',
          modeOff: 'Apagado',
          modeAuto: 'Ocultar automáticamente',
          modeAlways: 'Siempre visible',
          keyboardLedStatus: 'Indicadores de bloqueo del teclado',
          keyboardLedStatusDesc:
            'Mostrar el estado de Bloq Num, Bloq Mayús y Bloq Despl del equipo remoto',
          icons: 'Iconos del submenú',
          iconsDesc: 'Mostrar iconos de submenú en la barra de menú'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Estado de bloqueos del teclado remoto',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Bloq Num',
        numLockShort: 'Num',
        capsLock: 'Bloq Mayús',
        capsLockShort: 'May',
        scrollLock: 'Bloq Despl',
        scrollLockShort: 'Despl',
        on: 'Activado',
        off: 'Desactivado',
        unknown: 'Desconocido'
      },
      device: {
        title: 'Dispositivo',
        oled: {
          title: 'OLED',
          description: 'La pantalla OLED entra en reposo automáticamente',
          brightness: 'Brillo de la OLED',
          brightnessDescription: 'Un nivel más bajo alarga la vida de la pantalla',
          brightnessLevels: {
            '64': 'Mínimo',
            '96': 'Bajo',
            '128': 'Medio',
            '160': 'Alto',
            '207': 'Predeterminado',
            '255': 'Máximo'
          },
          0: 'Nunca',
          15: '15 s',
          30: '30 s',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 hora'
        },
        sections: {
          video: 'Vídeo',
          usb: 'USB',
          frontPanel: 'Panel frontal'
        },
        hidModeDesc:
          'Prueba el modo solo HID si el host no acepta el teclado y el ratón. Desactiva las unidades virtuales y la red.',
        resetHidDesc:
          'Vuelve a conectar el teclado y el ratón al host. Úsalo si la entrada deja de funcionar.',
        cpuFreq: {
          title: 'Frecuencia de la CPU',
          description: 'Establece la frecuencia de la CPU que se aplica en el próximo arranque',
          tip: 'La CPU arranca a 850 MHz y está especificada para 1000 MHz. Un valor nuevo se aplica en el próximo arranque, no mientras el sistema está en marcha. 1000 MHz está dentro de especificación; la temperatura queda muy por debajo de los límites con cualquiera de los dos ajustes.',
          running: 'En uso: {{mhz}} MHz',
          rebootToApply: 'reinicia para aplicar',
          rebootConfirm: '¿Reiniciar ahora para aplicar {{mhz}} MHz?'
        },
        swap: {
          title: 'Memoria Swap',
          disable: 'Desactivar',
          description: 'Establece el tamaño del archivo swap',
          tip: 'Habilitar esta función podría acortar la vida útil de tu tarjeta SD.'
        },
        zram: {
          title: 'Swap comprimida (zram)',
          description: 'Swap en RAM comprimida, en lugar de en la tarjeta SD',
          tip: 'zram mantiene la swap fuera de la tarjeta SD, así que no la desgasta. No hay swap en disco detrás: si zram se llena, el kernel detiene un proceso en lugar de paginar lentamente. El límite de memoria fija cuánta RAM puede ocupar zram.',
          unavailable: 'Los módulos del kernel no están instalados en este dispositivo',
          inactive: 'Activada, pero el dispositivo no se inició',
          active: 'Activa - {{used}} de {{total}}, {{ratio}}x',
          off: 'Desactivada',
          detail: {
            algorithm: 'Algoritmo: {{algorithm}}',
            memory: 'Memoria usada: {{used}} de {{limit}}',
            memoryNoLimit: 'Memoria usada: {{used}}, sin límite',
            counters:
              'Páginas intercambiadas: entrada {{in}}, salida {{out}} (todos los dispositivos swap, desde el arranque)'
          }
        },
        mouseJiggler: {
          title: 'Mouse Jiggler',
          description: 'Evitar que el host remoto entre en reposo',
          disable: 'Desactivar',
          absolute: 'Modo absoluto',
          relative: 'Modo relativo'
        },
        mdns: {
          description: 'Habilitar servicio de descubrimiento mDNS',
          tip: 'Desactívalo si no es necesario'
        },
        hdmi: {
          description: 'Habilitar salida HDMI/monitor',
          idleTimeoutTitle: 'Tiempo de espera de captura inactiva',
          idleTimeoutDescription:
            'Detener la captura HDMI después de no haber espectadores activos durante',
          minutes: 'min'
        },
        hidOnly: 'Modo sólo HID',
        hidOnlyDesc:
          'Dejar de emular dispositivos virtuales y conservar solo el control básico HID',
        disk: 'Disco Virtual',
        diskDesc: 'Montar disco virtual en el host remoto',
        network: 'Red Virtual',
        networkDesc: 'Montar tarjeta de red virtual en el host remoto',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Host:',
          description:
            'Un enlace de red privado con el host remoto a través del cable USB. El host recibe una dirección sin puerta de enlace ni DNS, así que no puede llegar a tu LAN a través de IronKVM.',
          off: 'Desactivado',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (para hosts sin NCM)',
          rndis: 'RNDIS (ya no se ofrece)',
          rndisNote: 'Este enlace usa RNDIS, que ya no se ofrece. Elige NCM o ECM.',
          subnet: 'Subred',
          subnetDesc:
            'Una red IPv4 privada, de /24 a /30. IronKVM toma la primera dirección y el host la segunda.',
          invalidSubnet: 'Introduce una subred como 172.31.255.0/30.',
          apply: 'Aplicar',
          confirm: '¿Reconectar el dispositivo USB?',
          reenumerate:
            'Al aplicar se reconstruye la conexión USB. El host pierde el teclado, el ratón y el disco virtual durante unos segundos.'
        },
        audio: 'Altavoz virtual',
        audioDesc:
          'Presenta una tarjeta de sonido USB al host remoto para que puedas oírlo. El host debe seleccionarla como dispositivo de salida. Cambiar esto reconstruye la conexión USB.',
        audioNote: 'El audio está disponible en ambos modos H.264 (WebRTC y Direct), no en MJPEG',
        console: 'Consola serie',
        consoleDesc:
          'Presenta un puerto serie USB al host remoto, para iniciar sesión en este IronKVM cuando la red no está disponible',
        consoleTip:
          'Quien controle el host remoto obtiene un aviso de inicio de sesión de este IronKVM. Establece una contraseña segura antes de habilitarlo (Cuenta - Cambiar contraseña).',
        endpoints: {
          title: 'Endpoints USB',
          used: '{{used}} de {{total}} en uso',
          cost: 'usa {{cost}}',
          needs: 'necesita {{cost}}',
          full: 'No hay suficientes endpoints USB. Desactiva otra cosa primero.',
          inactive:
            'Activado, pero sin funcionar: el controlador USB se quedó sin endpoints. Desactiva otro dispositivo y este arrancará de inmediato.',
          explain:
            'El controlador USB tiene un número fijo de endpoints de entrada, y esto los cuenta. Si se activan más dispositivos de los que caben, se conservan el teclado y el ratón y el resto se desactiva.',
          error: 'No se pudo contactar con el dispositivo. Inténtalo de nuevo.',
          fitTogether: 'Caben juntos: {{sets}}'
        },
        reboot: 'Reiniciar',
        rebootDesc: '¿Estás seguro de que deseas reiniciar el IronKVM?',
        okBtn: 'Sí',
        cancelBtn: 'No',
        rebootFailed: 'El reinicio ha fallado'
      },
      network: {
        title: 'Red',
        wifi: {
          disconnectBtn: 'Desconectar',
          disconnectWarning:
            'Si accedes a IronKVM por esta red Wi-Fi, esta página perderá la conexión.',
          disconnected: 'Wi-Fi desconectado',
          title: 'Wi-Fi',
          description: 'Configura el Wi-Fi',
          apMode: 'El modo AP está activado; conéctate al Wi-Fi escaneando el código QR',
          connect: 'Conectar Wi-Fi',
          connectDesc1: 'Introduce el SSID de la red y la contraseña',
          connectDesc2: 'Introduce la contraseña para unirte a esta red',
          disconnect: '¿Seguro que quieres desconectar la red?',
          failed: 'Error de conexión, inténtalo de nuevo.',
          ssid: 'Nombre',
          password: 'Contraseña',
          joinBtn: 'Unirse',
          confirmBtn: 'Aceptar',
          cancelBtn: 'Cancelar'
        },
        tls: {
          description: 'Habilitar protocolo HTTPS',
          tip: 'Aviso: Usar HTTPS puede aumentar la latencia, especialmente en modo de vídeo MJPEG.',
          restarting: 'Reiniciando el servidor del dispositivo, tardará unos dos minutos...',
          waiting: 'Esperando a que el dispositivo vuelva a responder...',
          waitingHttp: 'Volviendo a http. Recarga esta página si no se abre sola.',
          failed: 'No se pudo cambiar la configuración HTTPS',
          enableConfirm: '¿Activar HTTPS?',
          disableConfirm: '¿Desactivar HTTPS?',
          confirmDesc:
            'Esto cierra la sesión y reinicia el servidor del dispositivo, lo que tarda unos dos minutos. Después la página abre {{url}}.',
          confirmOk: 'Continuar',
          confirmCancel: 'Cancelar'
        },
        ethernet: {
          title: 'Dirección IP',
          description: 'Configure cómo IronKVM obtiene su dirección en la red cableada',
          dhcp: 'DHCP',
          manual: 'Manual',
          networkDetails: 'Detalles de red',
          interface: 'Interfaz',
          ipAddress: 'Dirección IP',
          subnetMask: 'Máscara de subred',
          router: 'Router',
          save: 'Aplicar',
          invalidAddress: 'Introduzca una dirección IP válida',
          invalidMask: 'Introduzca una máscara de subred válida, como 255.255.255.0 o 24',
          invalidRouter: 'Introduzca una dirección de router válida',
          addressRequired: 'Se requiere una dirección IP',
          maskRequired: 'Se requiere una máscara de subred',
          applyTitle: '¿Cambiar la dirección de IronKVM?',
          applyWarning:
            'Se perderá la conexión con esta página. IronKVM aplica la nueva dirección y espera {{seconds}} segundos a que llegue a él en esa dirección. Llegar a él conserva el cambio. Si no llega nada, IronKVM restaura la configuración anterior.',
          applyConfirm: 'Aplicar',
          applyCancel: 'Cancelar',
          applyFailed: 'No se pudo aplicar la dirección',
          trialTitle: 'Esperando confirmación',
          trialDhcp: 'IronKVM está solicitando una dirección por DHCP.',
          trialStatic: 'IronKVM está ahora en {{address}}.',
          trialInstruction:
            'Abra IronKVM en su nueva dirección e inicie sesión si se la pide. Llegar a él allí conserva el cambio. Si no llega nada a IronKVM en {{seconds}} segundos, restaura la configuración anterior.',
          trialOpen: 'Abrir la nueva dirección',
          trialKeep: 'Conservar esta configuración',
          trialKept: 'La nueva dirección está guardada',
          trialKeepFailed: 'No se pudo conservar la configuración',
          trialGone: 'El cambio ya se restauró. Inténtelo de nuevo.',
          unsaved: 'Cambios sin guardar'
        },
        dns: {
          title: 'DNS',
          description: 'Configura los servidores DNS para IronKVM',
          mode: 'Modo',
          dhcp: 'DHCP',
          manual: 'Manual',
          add: 'Añadir DNS',
          save: 'Guardar',
          invalid: 'Introduce una dirección IP válida',
          noDhcp: 'No hay DNS DHCP disponible actualmente',
          saved: 'Configuración DNS guardada',
          saveFailed: 'No se pudo guardar la configuración DNS',
          unsaved: 'Cambios sin guardar',
          maxServers: 'Se permiten como máximo {{count}} servidores DNS',
          dnsServers: 'Servidores DNS',
          dhcpServersDescription: 'Los servidores DNS se obtienen automáticamente por DHCP',
          manualServersDescription: 'Los servidores DNS se pueden editar manualmente',
          networkDetails: 'Detalles de red',
          interface: 'Interfaz',
          ipAddress: 'Dirección IP',
          subnetMask: 'Máscara de subred',
          router: 'Router',
          none: 'Ninguno'
        }
      },
      vpn: {
        connect: 'Conectar',
        connectDesc: 'Unirse a la red {{name}}. Apagado desconecta sin detener el servicio.',
        kvmUrl: 'Dirección del KVM',
        moreTip: 'Más acciones',
        restartTip: 'Reiniciar',
        stopTip: 'Detener',
        updateTip: 'Actualizar a {{version}}',
        loading: 'Cargando...',
        okBtn: 'Sí',
        cancelBtn: 'No',
        restart: '¿Reiniciar {{name}}?',
        stop: '¿Detener {{name}}?',
        stopDesc:
          'El daemon se detiene ahora. Iniciar al arrancar es un interruptor aparte y se queda como está.',
        update: '¿Actualizar {{name}} a {{version}}?',
        updateDesc: 'El daemon se reinicia si está en marcha. La sesión se conserva.',
        notInstall: '{{name}} no está instalado.',
        install: 'Instalar',
        installing: 'Instalando',
        installFailed: 'Error en la instalación',
        retry: 'Reintentar',
        notRunning: '{{name}} no está en ejecución. Inícialo para continuar.',
        run: 'Iniciar',
        boot: 'Iniciar al arrancar',
        bootDesc: 'Iniciar {{name}} cuando arranca el KVM.',
        control: 'Servidor de control',
        connected: 'Conectado',
        disconnected: 'No conectado',
        deviceName: 'Nombre del dispositivo',
        deviceIP: 'IP del dispositivo',
        account: 'Cuenta',
        version: 'Versión',
        uptime: 'Tiempo activo',
        peers: 'Pares',
        noPeers: 'Aún no hay pares.',
        online: 'En línea',
        offline: 'Sin conexión',
        memory: 'Memoria',
        daemonRss: 'Daemon',
        group: 'Grupo de complementos',
        high: 'se limita por encima de {{size}}',
        max: 'el kernel lo detiene por encima de {{size}}',
        noGroup: 'Esta placa no tiene grupo de memoria para complementos.',
        uninstall: 'Desinstalar {{name}}',
        uninstallDesc: '¿Seguro que quieres desinstalar {{name}}? La sesión se queda en la placa.',
        blocked:
          '{{other}} está en ejecución o se inicia al arrancar. Solo puede funcionar una VPN a la vez: detén {{other}} y desactiva antes su inicio al arrancar.',
        swap: {
          title: 'Memoria swap',
          tip: 'Si al daemon le falta memoria, prueba a activar la memoria swap. Se configura en "Ajustes > Rendimiento".'
        },
        copy: 'Copiar',
        copied: 'Enlace copiado',
        copyFailed: 'No se pudo copiar el enlace. Selecciónelo y cópielo a mano.',
        open: 'Abrir',
        checkAgain: 'Comprobar de nuevo',
        notSignedIn:
          'Aún no ha iniciado sesión. Termine de iniciarla en el enlace y vuelva a comprobar.',
        checkFailed: 'No se pudo comprobar el estado de la sesión',
        loginWaiting:
          'Esta página lo comprueba cada pocos segundos y continúa cuando haya iniciado sesión.',
        uninstallFailed: 'La desinstalación ha fallado',
        loginFailed: 'Error al iniciar sesión'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Descargar el',
        package: 'paquete de instalación',
        unzip: 'y descomprimirlo',
        notLogin:
          'El dispositivo aún no ha sido vinculado. Por favor, inicia sesión y vincula este dispositivo a tu cuenta.',
        urlPeriod: 'Esta URL es válida por 10 minutos',
        login: 'Iniciar sesión',
        logout: 'Cerrar sesión',
        logoutDesc: '¿Estás seguro de que deseas cerrar sesión?',
        manualIntro: 'O instálalo a mano por SSH:',
        copyBinaries: 'Copia tailscale y tailscaled en {{dir}} del IronKVM',
        linksFile: 'En el mismo directorio, crea un archivo llamado links con estas dos líneas:',
        rebootRefresh: 'Reinicia el IronKVM y luego actualiza esta página'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Este dispositivo aún no se ha unido a una red NetBird. Únete con una clave de configuración o inicia sesión con SSO.',
        setupKey: 'Clave de configuración',
        setupKeyPlaceholder: 'Pega una clave de configuración del panel de NetBird',
        join: 'Unirse',
        or: 'o',
        sso: 'Iniciar sesión con SSO',
        urlPeriod: 'Esta URL es válida durante 10 minutos',
        logout: 'Dar de baja',
        logoutDesc:
          'Dar de baja elimina este par de tu cuenta de NetBird y borra aquí su configuración. Para volver a unirse hace falta una clave de configuración o un inicio de sesión con SSO, y el par puede recibir una IP nueva. ¿Continuar?',
        joinFailed: 'No se pudo unir a la red'
      },
      update: {
        title: 'Buscar actualizaciones',
        queryFailed: 'Error al obtener la versión',
        updateFailed: 'La actualización falló. Por favor, inténtalo de nuevo.',
        isLatest: 'Ya tienes la última versión.',
        available: 'Hay una actualización disponible. ¿Estás seguro de que quieres actualizar?',
        updating: 'Actualización iniciada. Por favor, espera...',
        confirm: 'Confirmar',
        cancel: 'Cancelar',
        preview: 'Vista previa de actualizaciones',
        previewDesc: 'Accede anticipadamente a nuevas funciones y mejoras',
        previewTip:
          'Ten en cuenta que las versiones de vista previa pueden contener errores o funcionalidades incompletas',
        customServer: {
          title: 'Servidor de actualizaciones personalizado',
          desc: 'Buscar y descargar actualizaciones en línea desde un servidor especificado',
          invalidUrl:
            'Introduce un directorio de servidor HTTP o HTTPS válido, sin parámetros de consulta, fragmentos ni latest.json.',
          loadFailed: 'No se pudo cargar la configuración del servidor de actualizaciones.',
          saveFailed: 'No se pudo guardar la configuración del servidor de actualizaciones.',
          saved: 'Se ha guardado la configuración del servidor de actualizaciones.',
          save: 'Guardar',
          confirmTitle: '¿Usar un servidor de actualizaciones personalizado?',
          confirmDesc:
            'SHA-512 solo comprueba que el paquete coincide con el manifiesto proporcionado por este servidor. No demuestra que el paquete sea una versión oficial de IronKVM. Un servidor defectuoso o malicioso puede inutilizar el dispositivo, provocar la pérdida de datos o comprometer el sistema.',
          confirm: 'Usar de todos modos',
          useSipeed: 'Usar el servidor oficial de Sipeed',
          previewDisabled:
            'Las actualizaciones preliminares no están disponibles mientras esté activado un servidor de actualizaciones personalizado.'
        },
        offline: {
          chooseFile: 'Elegir archivo',
          installing: 'Subida completa. Instalando...',
          noFile: 'Ningún archivo elegido',
          title: 'Actualizaciones sin conexión',
          desc: 'Actualización a través del paquete de instalación local',
          upload: 'Subir',
          checksumPlaceholder: 'Suma de comprobación SHA-256 (opcional)',
          invalidChecksum:
            'La suma de comprobación SHA-256 debe contener 64 caracteres hexadecimales.',
          checksumMismatch:
            'La verificación SHA-256 ha fallado. Es posible que el paquete esté dañado.',
          invalidName:
            'Formato de nombre de archivo no válido. Descargue desde las versiones de GitHub.',
          updateFailed: 'La actualización falló. Por favor, inténtalo de nuevo.'
        },
        updateTo: 'Actualizar a {{version}}',
        updateConfirmDesc:
          'El dispositivo instala la actualización y reinicia su servidor. Esta página se recarga cuando el servidor vuelve.',
        releaseNotes: 'Notas de la versión'
      },
      account: {
        title: 'Cuenta',
        webAccount: 'Nombre de la cuenta web',
        role: 'Rol',
        roles: { admin: 'Administrador', user: 'Usuario' },
        password: 'Contraseña',
        updateBtn: 'Actualizar',
        logoutBtn: 'Cerrar sesión',
        logoutDesc: '¿Estás seguro de que deseas cerrar sesión?',
        okBtn: 'Sí',
        cancelBtn: 'No',
        users: {
          title: 'Usuarios',
          create: 'Crear usuario',
          enabled: 'Activado',
          disabled: 'Desactivado',
          deviceOwner: 'Propietario del dispositivo',
          resetPassword: 'Restablecer contraseña',
          delete: 'Eliminar',
          deleteConfirm: '¿Eliminar este usuario y revocar todas sus sesiones?',
          created: 'Usuario creado',
          deleted: 'Usuario eliminado',
          passwordUpdated: 'Contraseña actualizada',
          loadFailed: 'No se pudieron cargar los usuarios',
          saveFailed: 'No se pudo guardar el usuario',
          deleteFailed: 'No se pudo eliminar el usuario'
        }
      },
      apiKeys: {
        mcpNote: 'Estas claves no sirven para MCP, que tiene su propia clave en la página de MCP.',
        metricsUrl: 'URL de métricas',
        monitoring: 'Monitorización',
        monitoringDesc:
          'Prometheus lee las métricas con una clave API de esta página, enviada como token Bearer. Cualquier rol puede leerlas.',
        scrapeConfig: 'Configuración de scrape de Prometheus',
        title: 'Claves API',
        description:
          'Una clave actúa como su propietario, con el rol de ese usuario. Envíala como Authorization: Bearer <key> para las métricas y la API, o como X-Auth-Token para Redfish.',
        name: 'Nombre',
        namePlaceholder: 'Para qué es la clave, por ejemplo prometheus',
        nameRequired: 'Pon un nombre a la clave',
        nameTooLong: 'El nombre tiene como máximo 64 caracteres',
        unnamed: '(sin nombre)',
        create: 'Crear clave',
        created: 'Creada',
        owner: 'Propietario',
        empty: 'No hay claves API',
        newKeyTitle: 'Tu nueva clave API',
        newKeyWarning:
          'Copia la clave ahora. No se guarda y no se puede volver a mostrar. Si la pierdes, revócala y crea otra.',
        copy: 'Copiar',
        copied: 'Copiada',
        copyFailed: 'Error al copiar. Copia manualmente.',
        done: 'Listo',
        revoke: 'Revocar',
        revokeConfirmTitle: '¿Revocar esta clave API?',
        revokeConfirmDesc: 'Todo lo que use "{{name}}" dejará de funcionar al instante.',
        revoked: 'Clave API revocada',
        loadFailed: 'No se pudieron cargar las claves API',
        createFailed: 'No se pudo crear la clave API',
        revokeFailed: 'No se pudo revocar la clave API',
        cancelBtn: 'Cancelar'
      }
    },
    picoclaw: {
      title: 'PicoClaw Asistente',
      empty: 'Abre el panel e inicia una tarea para comenzar.',
      inputPlaceholder: 'Describe lo que quieres que haga el PicoClaw',
      newConversation: 'Nueva conversación',
      processing: 'Procesando...',
      agent: {
        defaultTitle: 'Asistente general',
        defaultDescription: 'Ayuda general para chat, búsqueda y espacio de trabajo.',
        kvmTitle: 'Control remoto',
        kvmDescription: 'Opere el host remoto a través de IronKVM.',
        switched: 'Rol de agente cambiado',
        switchFailed: 'No se pudo cambiar la función del agente'
      },
      send: 'Enviar',
      cancel: 'Cancelar',
      status: {
        connecting: 'Conectándose a la puerta de enlace...',
        connected: 'Sesión de PicoClaw conectada',
        disconnected: 'Sesión de PicoClaw cerrada',
        stopped: 'Solicitud de detención enviada',
        runtimeStarted: 'Tiempo de ejecución de PicoClaw iniciado',
        runtimeStartFailed: 'Error al iniciar el tiempo de ejecución de PicoClaw',
        runtimeStopped: 'Tiempo de ejecución de PicoClaw detenido',
        runtimeStopFailed: 'No se pudo detener el tiempo de ejecución de PicoClaw',
        controlSwitchedToMCP: 'El control se ha transferido al servicio MCP externo'
      },
      connection: {
        runtime: {
          checking: 'Comprobando',
          restoring: 'Restaurando PicoClaw',
          ready: 'Tiempo de ejecución listo',
          stopped: 'Tiempo de ejecución detenido',
          blockedByMCP: 'El control MCP externo está activo',
          readyBlockedByMCP:
            'El tiempo de ejecución está activo, pero un MCP externo controla ahora la entrada del dispositivo.',
          readyWithoutControl:
            'El tiempo de ejecución está activo. Concede a PicoClaw el control del dispositivo antes de volver a conectar.',
          unavailable: 'Tiempo de ejecución no disponible',
          configError: 'Error de configuración'
        },
        transport: {
          connecting: 'Conectando',
          connected: 'Conectado',
          disconnected: 'Desconectado',
          reconnect: 'Volver a conectar',
          reconnectDescription: 'Volver a conectar con la sesión de PicoClaw en curso.',
          reconnectBlocked:
            'PicoClaw necesita el control del dispositivo antes de volver a conectar.'
        },
        run: {
          idle: 'Inactivo',
          busy: 'Ocupado'
        }
      },
      message: {
        toolAction: 'Acción',
        observation: 'Observación',
        screenshot: 'Captura de pantalla'
      },
      overlay: {
        locked: 'PicoClaw está controlando el dispositivo. La entrada manual está en pausa.'
      },
      control: {
        picoclaw: 'Control del dispositivo: PicoClaw',
        picoclawDescription:
          'PicoClaw puede enviar entrada de teclado y ratón. La entrada manual puede quedar en pausa.',
        mcp: 'Control del dispositivo: MCP externo',
        mcpDescription:
          'El MCP externo puede escribir en el dispositivo. PicoClaw no tomará el control de la entrada.',
        off: 'Control del dispositivo: desactivado',
        offDescription:
          'La IA no enviará entrada de teclado ni de ratón. El control manual sigue disponible.',
        transitioning: 'Control del dispositivo: cambiando',
        transitioningDescription: 'El control del dispositivo se está sincronizando. Espera.',
        grant: 'Conceder control',
        release: 'Liberar',
        releasing: 'Liberando...',
        switching: 'Cambiando...',
        releasingLabel: 'Control del dispositivo: liberando',
        releasingDescription:
          'Se está devolviendo el control del dispositivo. PicoClaw ha detenido las escrituras en curso.',
        granted: 'Control de PicoClaw concedido',
        released: 'Control de PicoClaw liberado',
        grantFailed: 'No se pudo conceder el control de PicoClaw',
        releaseFailed: 'No se pudo liberar el control de PicoClaw',
        grantConfirmTitle: '¿Cambiar el control del dispositivo a PicoClaw?',
        grantConfirmDesc: 'Se interrumpirán las escrituras de dispositivo del MCP externo.'
      },
      install: {
        install: 'Instalar PicoClaw',
        installing: 'Instalando PicoClaw',
        success: 'PicoClaw instalado correctamente',
        failed: 'Error al instalar PicoClaw',
        uninstalling: 'Desinstalando el tiempo de ejecución...',
        uninstalled: 'El tiempo de ejecución se desinstaló exitosamente.',
        uninstallFailed: 'Falló la desinstalación.',
        requiredTitle: 'PicoClaw no está instalado',
        requiredDescription:
          'Instala PicoClaw antes de iniciar el tiempo de ejecución de PicoClaw.',
        progressDescription: 'PicoClaw se está descargando e instalando.',
        stages: {
          preparing: 'Preparando',
          downloading: 'Descargando',
          extracting: 'Extrayendo',
          verifying: 'Verificando',
          installing: 'Instalando',
          installed: 'Instalado',
          install_timeout: 'Tiempo de espera agotado',
          install_failed: 'Falló'
        }
      },
      model: {
        requiredTitle: 'Se requiere configuración del modelo',
        requiredDescription: 'Configure el modelo PicoClaw antes de usar el chat PicoClaw.',
        docsTitle: 'Guía de configuración',
        docsDesc: 'Modelos y protocolos compatibles',
        menuLabel: 'Configurar modelo',
        modelIdentifier: 'Identificador de modelo',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Clave API',
        apiKeyPlaceholder: 'Introduzca la clave API del modelo',
        save: 'Guardar',
        saving: 'Guardando',
        saved: 'Configuración del modelo guardada',
        saveFailed: 'No se pudo guardar la configuración del modelo',
        invalid: 'Se requieren el identificador del modelo, API Base URL y la clave API'
      },
      uninstall: {
        menuLabel: 'Desinstalar',
        confirmTitle: 'Desinstalar PicoClaw',
        confirmContent:
          '¿Está seguro de que desea desinstalar PicoClaw? Esto eliminará el ejecutable y todos los archivos de configuración.',
        confirmOk: 'Desinstalar',
        confirmCancel: 'Cancelar'
      },
      history: {
        title: 'Historial',
        loading: 'Cargando sesiones...',
        emptyTitle: 'Aún no hay historial',
        emptyDescription: 'Las sesiones PicoClaw anteriores aparecerán aquí.',
        loadFailed: 'Error al cargar el historial de sesiones',
        deleteFailed: 'No se pudo eliminar la sesión',
        deleteConfirmTitle: 'Eliminar sesión',
        deleteConfirmContent: '¿Está seguro de que desea eliminar "{{title}}"?',
        deleteConfirmOk: 'Eliminar',
        deleteConfirmCancel: 'Cancelar',
        messageCount_one: '{{count}} mensaje',
        messageCount_other: '{{count}} mensajes',
        messageCount: '{{count}} mensajes'
      },
      config: {
        startRuntime: 'Iniciar PicoClaw',
        stopRuntime: 'Detener PicoClaw'
      },
      start: {
        enableConfirmTitle: '¿Cambiar el control a PicoClaw?',
        enableConfirmDesc: 'Al iniciar PicoClaw se desactivará el servicio MCP externo.',
        enableConfirmOk: 'Iniciar PicoClaw',
        enableConfirmCancel: 'Cancelar',
        title: 'Iniciar PicoClaw',
        description:
          'Inicia el tiempo de ejecución para comenzar a utilizar el asistente PicoClaw.',
        switchFromMCP: 'Cambiar a PicoClaw e iniciar',
        takeoverAndStart: 'Tomar el control e iniciar'
      }
    },
    error: {
      title: 'Hemos encontrado un problema',
      refresh: 'Actualizar',
      panel: 'Esta parte de la página ha dejado de funcionar',
      retry: 'Reintentar'
    },
    fullscreen: {
      toggle: 'Activar/Desactivar pantalla completa'
    },
    input: {
      disconnected: 'El teclado y el ratón no están conectados',
      disconnectedTls:
        'El navegador ha rechazado la conexión segura que transporta el teclado y el ratón, algo que hace sin preguntar. El certificado que generó este dispositivo aún no es de confianza. Abre esta dirección en una pestaña nueva, acepta el certificado y recarga. Instalar el certificado es la solución fiable.',
      disconnectedNever:
        'No se pudo abrir la conexión que transporta el teclado y el ratón. El resto de la página funciona porque no la usa. Comprueba que nada entre tú y el dispositivo la esté bloqueando.',
      disconnectedDropped:
        'La conexión que transporta el teclado y el ratón se perdió y no se ha recuperado. Se reconecta sola tras un reinicio; si esto persiste, recarga la página.',
      hidDisabled: 'El HID está desactivado en este dispositivo (/boot/disable_hid).',
      keyFailed: 'No se pudo enviar la tecla.'
    },
    speaker: { title: 'Altavoz', unmute: 'Activar sonido', mute: 'Silenciar' },
    upstream: {
      check: 'Buscar actualizaciones',
      updateTo: 'Actualizar a {{version}}',
      confirm: '¿Actualizar {{name}} a {{version}}?',
      confirmDesc:
        'La nueva versión se descarga de GitHub y se comprueba con las sumas de verificación que publica. Si algo falla, se conserva la versión actual.',
      ok: 'Actualizar',
      upToDate: 'Actualizado',
      builtIn: 'integrada',
      checkFailed: 'No se pudieron buscar actualizaciones: {{error}}',
      unverifiable: 'La versión {{version}} no se ofrece: {{reason}}',
      inUse: 'No se puede actualizar ahora: {{reason}}',
      running: 'Actualizando a {{version}}...',
      done: '{{name}} actualizado a {{version}}',
      failed: 'La última actualización falló: {{error}}'
    },
    menu: {
      collapse: 'Colapsar menú',
      expand: 'Expandir menú',
      more: 'Más'
    },
    ion: {
      checking: 'Comprobando la memoria de vídeo antes de iniciar la transmisión...',
      warn: 'Queda poca memoria de vídeo. Un solo reinicio del servidor la agotaría. Reinicia cuando te venga bien.',
      criticalTitle: 'No hay suficiente memoria de vídeo para iniciar la transmisión',
      criticalBody:
        'Iniciar el vídeo agotaría la memoria reservada y detendría el servidor. Todas las demás funciones siguen funcionando, incluidos el control de encendido y el reinicio. Solo un reinicio del IronKVM recupera esta memoria.',
      criticalContinue: 'Iniciar el vídeo de todos modos',
      criticalReboot: 'Reiniciar IronKVM',
      criticalRebooting: 'Reiniciando...'
    }
  }
};

export default es;
