const ca = {
  translation: {
    feedback: {
      enabled: '{{name}} activat',
      disabled: '{{name}} desactivat',
      failed: 'La sol·licitud ha fallat. Torna-ho a provar.',
      network:
        "No s'ha pogut contactar amb el dispositiu. Comprova la connexió i torna-ho a provar.",
      saved: 'Desat',
      timeout: 'El dispositiu ha trigat massa a respondre. Torna-ho a provar.'
    },
    common: {
      copy: 'Copia',
      copied: 'Copiat',
      copyFailed: "No s'ha pogut copiar. Selecciona el text i copia'l a mà.",
      notUpdating: "Sense actualitzar: l'última consulta ha fallat.",
      off: 'Aturat',
      running: 'En marxa',
      save: 'Desa',
      cancel: 'Cancel·la',
      delete: 'Suprimeix',
      remove: 'Elimina'
    },
    head: {
      desktop: 'Escriptori remot',
      login: 'Inici de sessió',
      changePassword: 'Canviar contrasenya',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Contrasenya canviada. Inicia la sessió amb la contrasenya nova.',
      cookieRejected:
        "El navegador s'ha negat a desar la sessió. Una galeta que va deixar una sessió HTTPS anterior no es pot substituir per http sense xifrar. Esborreu les galetes d'aquesta adreça, o obriu una finestra privada, i torneu a iniciar la sessió.",
      login: 'Inici de sessió',
      placeholderUsername: "Nom d'usuari",
      placeholderPassword: 'Contrasenya',
      placeholderCurrentPassword: 'Contrasenya actual',
      placeholderPassword2: 'Torna a introduir la contrasenya',
      noEmptyUsername: "Cal introduir el nom d'usuari",
      noEmptyPassword: 'Cal introduir la contrasenya',
      passwordLength: 'La contrasenya ha de tenir entre 8 i 72 caràcters',
      noAccount:
        "No s'ha pogut obtenir la informació de l'usuari, actualitza la pàgina web o restableix la contrasenya",
      invalidUser: "Nom d'usuari o contrasenya invàlids",
      locked: 'Massa inicis de sessió, si us plau, torna-ho a provar més tard',
      globalLocked: 'Sistema sota protecció, torneu-ho a provar més tard',
      error: 'Error inesperat',
      invalidCurrentPassword: 'La contrasenya actual no és correcta',
      changePassword: 'Canviar la contrasenya',
      changePasswordDesc: 'Per a la seguretat del dispositiu, canvia la contrasenya!',
      differentPassword: 'Les contrasenyes no coincideixen',
      illegalUsername: "El nom d'usuari conté caràcters no permesos",
      illegalPassword: 'La contrasenya conté caràcters no permesos',
      forgetPassword: 'Has oblidat la contrasenya',
      ok: "D'acord",
      cancel: 'Cancel·la',
      loginButtonText: 'Inicia sessió',
      tips: {
        reset1:
          'Per restablir les contrasenyes, mantingues premut el botó BOOT del IronKVM durant 10 segons.',
        reset3: 'Compte web per defecte:',
        reset4: 'Compte SSH per defecte:',
        change1: 'Aquesta acció canviarà les següents contrasenyes:',
        change2: "Contrasenya d'inici de sessió web",
        change3: 'Contrasenya root del sistema (inici de sessió SSH)',
        change4: 'Per restablir les contrasenyes, mantingues premut el botó BOOT del IronKVM.',
        resetDocs: 'Per als passos detallats, consulta la documentació del maquinari:',
        hardwareDocs: 'Wiki del Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Configura Wi-Fi per al IronKVM',
      success: "Comprova l'estat de la xarxa del IronKVM i visita la nova adreça IP.",
      failed: "L'operació ha fallat, torna-ho a intentar.",
      invalidMode:
        'El mode actual no admet la configuració de xarxa. Aneu al vostre dispositiu i activeu el mode de configuració Wi-Fi.',
      confirmBtn: "D'acord",
      finishBtn: 'Fet',
      ap: {
        authTitle: 'Es requereix autenticació',
        authDescription: 'Introduïu la contrasenya AP per continuar',
        authFailed: 'Contrasenya AP no vàlida',
        passPlaceholder: 'AP contrasenya',
        verifyBtn: 'Verificar'
      },
      ssidRequired: 'Introduïu el nom de la xarxa, fins a 32 caràcters',
      passwordLength:
        'La contrasenya té de 8 a 63 caràcters. Deixeu-la buida per a una xarxa oberta.',
      passwordOptional: 'Contrasenya (buida per a una xarxa oberta)',
      lost: "La placa ha deixat de respondre. Potser s'ha connectat a la xarxa i ha tancat el punt d'accés de configuració. Si el punt d'accés torna a aparèixer, la connexió ha fallat: torneu-vos-hi a connectar i torneu-ho a provar.",
      done: 'Configuració acabada. Torneu a connectar aquest dispositiu a la vostra xarxa habitual i obriu la placa a la seva nova adreça.'
    },
    screen: {
      viewOnly: 'Només veure',
      viewOnlyTip:
        "Aquesta pestanya deixa d'enviar teclat i ratolí a l'amfitrió. Els scripts, el jiggler del ratolí i altres espectadors no es veuen afectats.",
      viewOnlyOff: 'Desactiva només veure',
      viewOnlyBlocked: "Només veure és actiu, no s'ha enviat res a l'amfitrió",
      pauseHidden: 'Pausa amb la pestanya amagada',
      pauseHiddenTip:
        "Atura el vídeo i el so uns segons després d'amagar aquesta pestanya, i els reprèn quan hi tornes.",
      screenshot: 'Captura de pantalla',
      screenshotTip: "Desa la pantalla de l'amfitrió com a PNG a mida completa de captura.",
      screenshotFailed: 'Ha fallat la captura de pantalla',
      stream: {
        ok: 'imatge correcta',
        noSignal: 'sense senyal',
        failed: 'error del flux'
      },
      codecNoWebrtcHevc: 'Aquest navegador no pot rebre H.265 per WebRTC',
      codecNoHevc: 'Aquest navegador no pot descodificar H.265',
      codecNote:
        'La placa té un sol codificador, així que això canvia el flux per a tots els espectadors. Torneu a connectar per aplicar-ho a una sessió WebRTC en curs.',
      codec: 'Còdec',
      updateFailed: "No s'ha aplicat la configuració",
      scale: 'Escala',
      title: 'Pantalla',
      video: 'Mode de vídeo',
      videoDirectTips: "Activa HTTPS a 'Configuració > Dispositiu' per utilitzar aquest mode",
      resolution: 'Resolució',
      ocr: {
        title: 'Llegeix text (OCR)',
        tips: 'El text es reconeix en aquest navegador. El pots corregir abans de copiar-lo.',
        hint: 'Arrossega sobre el text que vols llegir. Prem Esc per cancel·lar.',
        noPicture: 'Espera el vídeo i després arrossega sobre el text que vols llegir.',
        cancel: 'Cancel·la',
        language: 'Idioma',
        languages: {
          eng: 'Anglès'
        },
        preview: 'Àrea seleccionada',
        capturing: "S'està capturant la pantalla...",
        loading: "S'està carregant el reconeixement de text...",
        recognizing: "S'està llegint el text...",
        noText: "No s'ha trobat cap text a l'àrea seleccionada.",
        copy: 'Copia',
        copied: 'Copiat al porta-retalls',
        copyFailed: "No s'ha pogut copiar al porta-retalls",
        selectAgain: 'Torna a seleccionar',
        unsupported:
          'Aquest navegador no pot executar el reconeixement de text. Necessita WebAssembly SIMD, que tenen els navegadors actuals.',
        captureFailed: "No s'ha pogut capturar la pantalla.",
        outside: "L'àrea seleccionada és fora de la imatge.",
        recognizeFailed: 'Ha fallat el reconeixement de text.'
      },
      controlRegion: {
        title: 'Calibratge del ratolí',
        description:
          'Utilitzeu aquesta opció quan el dispositiu controlat tingui una resolució que no sigui 16:9 i el cursor estigui desalineat horitzontalment o verticalment.',
        off: 'Desactivat',
        auto: 'Automàtic',
        autoWarning:
          "El calibratge pot fallar quan l'aplicació de l'usuari tingui un fons completament negre.",
        manual: 'Manual',
        selectedResolution: "Resolució de l'àrea seleccionada",
        unused: 'No utilitzada',
        originalResolution: 'Resolució original',
        selectResolution: 'Selecciona la resolució original',
        addResolution: 'Afegeix una resolució personalitzada',
        add: 'Afegeix',
        duplicateResolution: 'Aquesta resolució ja existeix.',
        width: 'Amplada',
        height: 'Alçada',
        apply: 'Calcula i aplica',
        invalidResolution: 'Introduïu una resolució original vàlida quan el vídeo estigui llest.',
        select: "Selecciona l'àrea",
        clear: 'Restaura la detecció automàtica',
        saveFailed: "No s'ha pogut desar l'àrea d'entrada.",
        tooSmall: "L'àrea seleccionada és massa petita.",
        previewUnavailable: 'Vista prèvia no disponible',
        clearConfirm: 'Voleu restaurar la detecció automàtica de vores negres?',
        dragHint: "Arrossegueu per seleccionar l'àrea de l'escriptori remot",
        finish: 'Fet',
        confirm: 'Confirma',
        cancel: 'Cancel·la'
      },
      auto: 'Automàtic',
      autoTips:
        "Poden aparèixer talls o desajustos del ratolí en certes resolucions. Prova a canviar la resolució de l'amfitrió remot o desactiva el mode automàtic.",
      fps: 'FPS',
      customizeFps: 'Personalitzat',
      quality: 'Qualitat',
      qualityLossless: 'Màxima',
      qualityHigh: 'Alta',
      qualityMedium: 'Mitjana',
      qualityLow: 'Baixa',
      frameDetect: 'Detecció de fotogrames',
      frameDetectTip:
        "Calcula la diferència entre fotogrames. S'atura la transmissió si no hi ha canvis a la pantalla de l'amfitrió remot.",
      resetHdmi: 'Restablir HDMI',
      mixedH264: {
        title: 'Conflicte de flux H.264',
        description:
          "S'estan utilitzant H.264 Direct i H.264 WebRTC alhora. Això pot provocar esquinçament de pantalla o vídeo corrupte. Utilitzeu només un mode H.264."
      },
      webrtcConnectionFailed: {
        title: 'Error de connexió WebRTC',
        description: 'Comproveu la connexió de xarxa o canvieu el mode de vídeo.'
      },
      captureStatus: {
        hdmiError: 'Error a la pantalla HDMI',
        unsupportedResolution: 'La resolució actual no és compatible',
        retrieving: "S'està obtenint la pantalla...",
        changingResolution: "S'està canviant la resolució...",
        updateFailed: 'La pantalla no es pot actualitzar ara',
        videoError: 'Error de visualització de vídeo',
        noHdmi: "No s'ha detectat cap senyal HDMI",
        unavailable: 'La pantalla no es pot mostrar ara'
      },
      directConnectionFailed: 'Ha fallat la connexió del flux de vídeo'
    },
    keyboard: {
      close: 'Tanca',
      title: 'Teclat',
      paste: 'Enganxa',
      tips: "Escriu el text a l'amfitrió com a pulsacions de tecles. Tria la distribució de teclat que fa servir l'amfitrió.",
      placeholder: 'Escriu aquí',
      submit: 'Envia',
      virtual: 'Teclat',
      readClipboard: 'Llegir des del porta-retalls',
      clipboardPermissionDenied:
        "S'ha denegat el permís del porta-retalls. Permet l'accés al porta-retalls al teu navegador.",
      clipboardReadError: "No s'ha pogut llegir el porta-retalls",
      mediaKeys: {
        title: 'Tecles multimèdia',
        mute: 'Silenci',
        volumeDown: 'Baixa el volum',
        volumeUp: 'Puja el volum',
        previous: 'Pista anterior',
        playPause: 'Reprodueix o posa en pausa',
        next: 'Pista següent',
        stop: 'Atura'
      },
      pasting: {
        layout: "Distribució de teclat de l'amfitrió",
        layouts: {
          us: 'Anglès (EUA)',
          uk: 'Anglès (Regne Unit)',
          de: 'Alemany',
          fr: 'Francès',
          es: 'Espanyol',
          it: 'Italià',
          ptBr: 'Portuguès (Brasil)',
          se: 'Suec / finès',
          ru: 'Rus',
          ja: 'Japonès',
          ko: 'Coreà'
        },
        speed: "Velocitat d'escriptura",
        speeds: {
          fast: 'Ràpida',
          normal: 'Normal',
          slow: 'Lenta'
        },
        estimate: "Temps d'escriptura: uns {{duration}}",
        untypeable: 'Caràcters que aquesta distribució no pot escriure: {{count}}',
        untypeableAt: 'línia {{line}}, columna {{column}}',
        skipUntypeable: 'Escriu la resta',
        shortcut: "{{shortcut}} escriu el porta-retalls a l'amfitrió directament.",
        clipboardUnavailable:
          'El navegador només deixa que una pàgina llegeixi el porta-retalls per HTTPS. Enganxa el text al quadre amb Ctrl+V.',
        clipboardEmpty: 'El porta-retalls no conté text.',
        tooLong: 'El text és massa llarg. El límit és de {{max}} caràcters.',
        inProgress: "Ja s'està escrivint un text enganxat.",
        typing: "Escrivint a l'amfitrió",
        done: 'Text escrit',
        canceled: 'Enganxament cancel·lat',
        failed: "L'enganxament ha fallat",
        cancel: 'Cancel·la',
        controlBusy: 'Un altre controlador està fent servir el teclat.',
        hidError: "No s'han pogut enviar les pulsacions a l'amfitrió."
      },
      shortcut: {
        sendFailed: "No s'ha enviat: la connexió d'entrada no està disponible",
        title: 'Dreceres',
        custom: 'Personalitzat',
        capture: 'Feu clic aquí per capturar la drecera',
        clear: 'Clar',
        save: 'Desa',
        captureTips:
          'Capturar tecles del sistema (com la tecla Windows) requereix permís de pantalla completa.',
        enterFullScreen: 'Canvia el mode de pantalla completa.'
      },
      leaderKey: {
        saveFailed: "No s'ha pogut desar la tecla líder",
        title: 'Tecla líder',
        desc: "Evita les restriccions del navegador i envia dreceres del sistema directament a l'amfitrió remot.",
        howToUse: "Com s'utilitza",
        simultaneous: {
          title: 'Mode simultània',
          desc1: 'Manteniu premuda la tecla líder i premeu la drecera.',
          desc2: 'Intuïtiu, però pot entrar en conflicte amb les dreceres del sistema.'
        },
        sequential: {
          title: 'Mode seqüencial',
          desc1:
            'Premeu la tecla líder → premeu la drecera en seqüència → torneu a prémer la tecla líder.',
          desc2: 'Requereix més passos, però evita completament els conflictes del sistema.'
        },
        enable: 'Activa la tecla líder',
        tip: "Quan s'assigna com a tecla líder, aquesta tecla només funciona com a activador de dreceres i perd el seu comportament predeterminat.",
        placeholder: 'Premeu la tecla líder',
        shiftRight: 'Maj dreta',
        ctrlRight: 'Ctrl dret',
        metaRight: 'Win dret',
        submit: 'Envia',
        recorder: {
          rec: 'REC',
          activate: 'Activa les tecles',
          input: 'Premeu la drecera...'
        }
      }
    },
    mouse: {
      jiggler: 'Moviment automàtic del ratolí',
      title: 'Ratolí',
      cursor: 'Estil del cursor',
      default: 'Cursor per defecte',
      pointer: 'Cursor punter',
      cell: 'Cursor de cel·la',
      text: 'Cursor de text',
      grab: 'Cursor de mà',
      hide: 'Amaga el cursor',
      mode: 'Mode de ratolí',
      absolute: 'Mode absolut',
      relative: 'Mode relatiu',
      absoluteShort: 'Absolut',
      relativeShort: 'Relatiu',
      touch: 'Mode tàctil',
      touchShort: 'Tàctil',
      absoluteStalled: "L'amfitrió ignora el ratolí absolut",
      absoluteStalledDesc:
        "L'amfitrió ha deixat de recollir els informes del ratolí absolut, de manera que els moviments del punter es perden. El teclat no se'n veu afectat. Recuperar l'USB sol resoldre-ho; el mode relatiu fa servir un altre endpoint.",
      useRelative: 'Canvia al mode relatiu',
      direction: 'Direcció de la roda de desplaçament',
      scrollUp: 'Igual que aquest ordinador',
      scrollDown: 'Invertit (desplaçament natural)',
      speed: 'Velocitat de desplaçament',
      fast: 'Ràpida',
      slow: 'Lenta',
      requestPointer: "Estàs usant el mode relatiu. Fes clic a l'escriptori per obtenir el punter.",
      resetHid: 'Restablir HID',
      hidOnly: {
        switchFailed: "No s'ha pogut canviar el mode. Comproveu la connexió i torneu-ho a provar.",
        title: 'Mode només HID',
        desc: 'Si el ratolí i el teclat deixen de respondre i restablir HID no ajuda, pot ser un problema de compatibilitat entre el IronKVM i el dispositiu. Proveu d’activar el mode només HID per millorar la compatibilitat.',
        tip1: 'Activar el mode només HID desmuntarà el disc virtual i la xarxa virtual',
        tip2: 'En mode només HID, no es pot muntar imatges',
        rebuild: 'Canviar de mode reconstrueix la connexió USB. El IronKVM no es reinicia',
        enable: 'Activa mode només HID',
        disable: 'Desactiva mode només HID'
      },
      resetHidDone: "S'ha reiniciat l'HID USB",
      resetHidFailed: "No s'ha pogut reiniciar l'HID USB"
    },
    image: {
      driveLoaded: 'imatge carregada',
      driveWarning: 'revisa els avisos',
      warning: {
        missing:
          "El fitxer d'imatge s'ha esborrat. L'amfitrió continua llegint la còpia antiga fins que l'expulsis.",
        writable: "Lectura i escriptura: l'amfitrió pot modificar aquesta imatge.",
        tooBigForCd:
          'Massa gran per a la unitat de CD ({{size}}, límit {{max}}). Fes servir el disc.',
        tooSmallForCd: 'Massa petita per a la unitat de CD ({{size}}). Fes servir el disc.',
        empty: 'El fitxer és buit, probablement per una pujada o baixada fallida.'
      },
      delete: 'Elimina',
      inUse: "En ús. Expulseu-la abans d'eliminar-la.",
      retry: 'Torna-ho a provar',
      loadFailed: "No s'ha pogut carregar la llista d'imatges",
      readOnlyLocked: "Expulseu el disc per canviar-ho. S'aplica quan s'insereix una imatge.",
      title: 'Imatges',
      loading: 'Carregant...',
      empty: "No s'ha trobat res",
      mountMode: 'Mode de muntatge',
      mountFailed: 'Error en muntar',
      mountDesc: 'En alguns sistemes cal expulsar el disc virtual abans de muntar la imatge.',
      unmountFailed: "No s'ha pogut desmuntar",
      unmountDesc:
        "En alguns sistemes, cal expulsar manualment de l'amfitrió remot abans de desmuntar la imatge.",
      refresh: 'Actualitza la llista',
      disk: 'Disc',
      cdrom: 'CD',
      driveEmpty: 'Buida',
      eject: 'Expulsa',
      readOnly: 'Només lectura',
      readOnlyTip: "S'aplica a la propera imatge que s'insereixi al disc.",
      noDrives: 'No hi ha unitats virtuals. Activeu el disc virtual a Configuració.',
      insertFailed: "No s'ha pogut inserir",
      ejectFailed: "No s'ha pogut expulsar",
      insertInto: 'Insereix a {{drive}}. Feu clic per canviar-ho.',
      loadedIn: 'A la unitat {{drive}}',
      attention: 'Atenció',
      deleteConfirm: 'Esteu segur que voleu suprimir aquesta imatge?',
      okBtn: 'Sí',
      cancelBtn: 'No',
      deleteFailed: 'Error en eliminar',
      ventoy: {
        statusNoKernel: 'No compatible amb aquest firmware',
        statusNotInstalled: 'No instal·lat',
        statusReady: 'A punt',
        statusSelected: 'Imatges seleccionades: {{count}}',
        statusInDrive: 'A la unitat de disc, {{size}}',
        noKernel:
          "El nucli d'aquest firmware no té suport de device-mapper, així que Ventoy no es pot fer servir fins que s'instal·li una imatge que el tingui.",
        installDesc: "Arrenca l'amfitrió des de diverses imatges en un sol disc, sense copiar-les.",
        install: 'Instal·la',
        installing: "S'està baixant Ventoy, uns 20 MB. Pot trigar uns minuts.",
        needsData: "Ventoy necessita una imatge d'IronKVM amb la partició /data muntada.",
        uninstall: 'Desinstal·la',
        uninstallConfirm: 'Voleu eliminar els fitxers de Ventoy?',
        noImages: 'No hi ha imatges per posar al disc Ventoy.',
        onDisk: 'Al disc Ventoy',
        missing: 'Falta: {{file}}',
        remove: 'Treu del disc Ventoy',
        setHint: "El conjunt d'imatges només canvia mentre el disc Ventoy no és en cap unitat.",
        useAsDisk: 'Fes servir com a disc virtual',
        failed: 'La sol·licitud de Ventoy ha fallat',
        secureBoot:
          "Amb Secure Boot activat, l'amfitrió ha de registrar una vegada la clau de Ventoy a MokManager. El fitxer de clau ENROLL_THIS_KEY_IN_MOKMANAGER.cer és a la partició VTOYEFI.",
        readOnly:
          "L'amfitrió veu el disc només de lectura, així que la persistència de Ventoy i ventoy.json a la unitat no funcionen."
      },
      tips: {
        title: 'Com pujar imatges',
        usb1: 'Connecta el IronKVM al teu ordinador via USB.',
        usb2: "Assegura't que el disc virtual està muntat (Configuració - Disc Virtual).",
        usb3: "Obre el disc virtual i copia la imatge a l'arrel.",
        scp1: 'Comprova que el IronKVM i el teu ordinador estan a la mateixa xarxa.',
        scp2: 'Obre un terminal i usa SCP per pujar la imatge al directori /data del IronKVM.',
        scp3: 'Exemple: scp ruta-de-la-imatge root@ip-del-nanokvm:/data',
        tfCard: 'Targeta TF',
        tf1: 'Mètode disponible a sistemes GNU/Linux',
        tf2: 'Extreu la targeta TF del IronKVM (versió FULL, cal obrir la carcassa).',
        tf3: "Introdueix la targeta en un lector i connecta'l a l'ordinador.",
        tf4: 'Copia la imatge al directori /data de la targeta.',
        tf5: 'Reintrodueix la targeta al IronKVM.'
      }
    },
    script: {
      title: 'Scripts',
      upload: 'Puja',
      run: 'Executa',
      runBackground: 'Executa en segon pla',
      runFailed: "Error en l'execució",
      attention: 'Atenció',
      delDesc: 'Estàs segur que vols eliminar aquest fitxer?',
      confirm: 'Sí',
      cancel: 'No',
      delete: 'Esborra',
      close: 'Tanca',
      empty: 'Encara no hi ha scripts. Pugeu un fitxer .sh o .py per executar-lo a la placa.',
      loadFailed: "No s'han pogut carregar els scripts",
      uploaded: 'Script pujat',
      uploadFailed: "No s'ha pogut pujar l'script",
      started: 'Script iniciat en segon pla',
      deleteFailed: "No s'ha pogut suprimir l'script",
      waitLimit: "S'espera que acabi l'script, fins a {{minutes}} minuts.",
      timedOut:
        "L'script ha trigat més de {{minutes}} minuts i aquesta pàgina ha deixat d'esperar. Potser encara s'està executant a la placa."
    },
    terminal: {
      invalidBaud: 'Aquesta velocitat en bauds no és compatible.',
      invalidPort: 'Introduïu un camí de dispositiu a /dev, com ara /dev/ttyS1.',
      invalidSettings: 'Configuració del port sèrie no vàlida. Aquest és el terminal de la placa.',
      disconnected: 'Desconnectat. Premeu Retorn per tornar a connectar.',
      title: 'Terminal',
      nanokvm: 'Terminal IronKVM',
      serial: 'Terminal de port sèrie',
      serialPort: 'Port sèrie',
      serialPortPlaceholder: 'Introdueix el port sèrie',
      baudrate: 'Velocitat (baudrate)',
      parity: 'Paritat',
      parityNone: 'Cap',
      parityEven: 'Parell',
      parityOdd: 'Senar',
      flowControl: 'Control de flux',
      flowControlNone: 'Cap',
      flowControlSoft: 'Programari',
      flowControlHard: 'Maquinari',
      dataBits: 'Bits de dades',
      stopBits: 'Bits de parada',
      confirm: "D'acord"
    },
    wol: {
      no: 'No',
      yes: 'Sí',
      deleteConfirm: 'Voleu eliminar aquesta adreça desada?',
      delete: 'Elimina',
      wake: 'Desperta',
      rename: 'Canvia el nom',
      showMac: "Mostra l'adreça MAC",
      showName: 'Mostra el nom',
      requestFailed: "No s'ha pogut contactar amb el dispositiu per enviar l'ordre",
      deleteFailed: "No s'ha pogut eliminar",
      renameFailed: "No s'ha pogut canviar el nom",
      title: 'Wake-on-LAN',
      sending: 'Enviant comanda...',
      sent: 'Comanda enviada',
      input: 'Introdueix la MAC',
      ok: "D'acord"
    },
    download: {
      uploadFailed: 'La pujada ha fallat',
      uploadSuccess: 'Pujada completada',
      uploading: "S'està pujant: {{file}}",
      downloadingPercent: "S'està baixant ({{percent}}): {{file}}",
      downloading: "S'està baixant: {{file}}",
      title: "Descarregador d'imatges",
      input: 'Introdueix la URL de la imatge',
      ok: "D'acord",
      disabled: 'La partició /data és només lectura. No es pot descarregar la imatge.',
      uploadbox: 'Deixeu anar el fitxer aquí o feu clic per seleccionar-lo',
      inputfile: "Introduïu el fitxer d'imatge",
      NoISO: 'Cap ISO',
      sha256: 'SHA-256 (opcional)',
      sha256Placeholder: 'Introduïu una suma de verificació SHA-256 de 64 caràcters',
      invalidSHA256: 'SHA-256 ha de ser una cadena hexadecimal de 64 caràcters',
      failed: 'Descàrrega fallida',
      success: 'Descàrrega correcta',
      checksumFailed: 'Descàrrega fallida: ha fallat la verificació SHA-256',
      cancel: 'Cancel·la',
      cancelFailed: 'No sha pogut cancel·lar la descàrrega',
      bootMenu: "Menú d'arrencada (netboot.xyz)",
      bootMenuPresent: '{{file}} ja és a la placa, amb la suma correcta',
      bootMenuDesc: 'Baixa la ISO de netboot.xyz, amb la suma comprovada, per al CD virtual'
    },
    alerts: {
      title: 'Cal atenció',
      temperature: {
        warning: 'La placa és a {{celsius}} °C. Comprova que li arribi aire.',
        critical: 'La placa és a {{celsius}} °C, massa calenta. Dona-li aire o apaga-la.'
      },
      storage: {
        warning:
          'Només queden {{available}} lliures de {{total}} a {{path}}. Les imatges grans potser no hi cabran.',
        critical:
          'Només queden {{available}} lliures a {{path}}. Fallaran pujades, baixades i instal·lacions de complements. Esborra les imatges que ja no necessitis.'
      },
      vpn: "{{name}} s'ha d'iniciar en arrencar però no s'executa, així que l'accés remot a través seu no funciona.",
      openVpn: 'Obre la configuració de VPN',
      stream:
        'El flux de vídeo ha fallat. Prova un altre mode de vídeo al menú Pantalla o torna a carregar la pàgina.'
    },
    power: {
      resetDesc: "Reinicia l'amfitrió a l'instant. Es perd la feina no desada.",
      powerShortDesc: "Engega l'amfitrió o demana al seu sistema operatiu que s'apagui (ACPI).",
      powerLongDesc: "Força l'apagada de l'amfitrió sense tancar el sistema.",
      hddLed: 'LED de disc',
      hddActive: 'Actiu',
      hddIdle: 'Inactiu',
      title: 'Alimentació',
      showConfirm: 'Confirmació',
      showConfirmTip:
        "Demana confirmació abans d'una pulsació curta. El reinici i la pulsació llarga sempre la demanen.",
      reset: 'Reinicia',
      power: 'Encén',
      powerShort: 'Clic curt',
      powerLong: 'Clic llarg',
      resetConfirm: 'Vols realment reiniciar?',
      powerConfirm: 'Vols realment encendre/apagar?',
      okBtn: 'Sí',
      cancelBtn: 'No',
      hostOs: "SO de l'amfitrió",
      hostOsTip: "S'envien com a tecles USB. L'amfitrió decideix què fan.",
      sleep: 'Suspèn',
      wake: 'Desperta',
      wakeKey: 'Desperta amb Maj',
      powerDown: 'Apaga',
      sleepConfirm: "Vols suspendre l'amfitrió?",
      powerDownConfirm: "Vols enviar la tecla d'apagada a l'amfitrió?",
      wakeTip:
        "Un amfitrió suspès sovint ignora Desperta del dispositiu que l'ha suspès. Desperta amb Maj prem una tecla del teclat, que més amfitrions accepten.",
      led: "LED d'alimentació",
      ledOn: 'Encès',
      ledOff: 'Apagat',
      ledUnknown: 'Desconegut',
      ledConnected: "LED d'alimentació connectat",
      ledConnectedTip:
        "Activeu-ho només si el connector del LED d'alimentació de l'amfitrió està cablejat a la placa. Sense això, l'estat d'alimentació és desconegut.",
      ledConnectedFailed: "No s'ha pogut desar la configuració del LED d'alimentació",
      powerLongConfirm:
        "Mantenir premut el botó d'engegada {{seconds}} s? Això talla l'alimentació sense apagar el sistema.",
      done: 'Botó premut',
      failed: "No s'ha pogut prémer el botó"
    },
    settings: {
      title: 'Configuració',
      nav: {
        system: 'Sistema',
        network: 'Xarxa',
        access: 'Accés',
        integrations: 'Integracions',
        boot: 'Arrencada i mitjans',
        browser: 'Aquest navegador',
        search: 'Cerca un ajust',
        noMatch: 'Cap ajust coincideix',
        locked:
          'Hi ha una operació en curs. Les altres pàgines i el tancament no estan disponibles fins que acabi.',
        vpnProvider: 'Proveïdor de VPN'
      },
      mcp: {
        keyNote:
          'MCP fa servir la seva pròpia clau API, que es mostra a sota. Les claus de la pàgina Claus API no funcionen aquí.',
        title: 'Servei MCP',
        service: 'Control remot MCP',
        serviceDesc:
          'Permet que clients MCP de confiança controlin el teclat i el ratolí i capturin pantalles',
        securityWarning:
          'Qualsevol persona amb aquesta clau API pot controlar l’amfitrió remot i veure’n la pantalla. Utilitzeu HTTPS i activeu-lo només en xarxes de confiança.',
        endpoint: 'Punt de connexió',
        apiKey: 'Clau API',
        regenerateConfirmTitle: 'Voleu tornar a generar la clau API MCP?',
        regenerateConfirmDesc: 'La clau actual deixarà de funcionar immediatament.',
        enableConfirmTitle: 'Voleu activar el control MCP extern?',
        enableConfirmDesc:
          'En activar MCP, PicoClaw s’aturarà i es tancarà qualsevol sessió activa de PicoClaw.',
        failed: 'L’operació MCP ha fallat',
        copyFailed: 'La còpia ha fallat. Copieu-ho manualment.',
        okBtn: 'Confirma',
        cancelBtn: 'Cancel·la',
        showKey: 'Mostra la clau',
        hideKey: 'Amaga la clau',
        regenerateKey: 'Regenera la clau'
      },
      redfish: {
        example: 'Exemple',
        title: 'Redfish',
        service: 'Servei Redfish',
        serviceDesc:
          "L'API Redfish de la DMTF, per al control d'alimentació, els suports virtuals i l'estat des d'eines com redfishtool i Ansible. Desactivar-la tanca totes les sessions Redfish.",
        endpoint: 'Arrel del servei',
        httpsOn: "La placa serveix HTTPS, que la majoria d'eines Redfish necessiten.",
        httpsOff:
          'La placa serveix HTTP sense xifrar. La majoria d\'eines Redfish necessiten HTTPS: activeu-lo a "Configuració > Xarxa".',
        credentials:
          'Redfish accepta els comptes del KVM, amb autenticació Basic o una sessió Redfish, i claus API enviades com a X-Auth-Token. Les claus API es gestionen a la pàgina Claus API.',
        powerActions: "Accions d'alimentació",
        powerActionsDesc:
          "Els tipus de reinici que s'ofereixen ara. On, ForceOff i GracefulShutdown necessiten l'estat d'alimentació, així que només s'ofereixen quan \"LED d'alimentació connectat\" està activat al menú d'alimentació.",
        sessions: 'Sessions',
        noSessions: 'No hi ha cap sessió Redfish oberta',
        created: 'Creada',
        lastUsed: 'Últim ús',
        refresh: 'Actualitza',
        end: 'Tanca',
        endConfirmTitle: 'Voleu tancar aquesta sessió Redfish?',
        endConfirmDesc:
          'El seu testimoni deixa de funcionar immediatament. El client haurà de tornar a iniciar la sessió.',
        failed: "L'operació Redfish ha fallat",
        copyFailed: 'La còpia ha fallat. Copieu-ho manualment.',
        okBtn: 'Confirma',
        cancelBtn: 'Cancel·la'
      },
      ipmi: {
        copyBeforeSave: 'Copia la contrasenya ara. Un cop desada, no es pot tornar a mostrar.',
        noLogin:
          'IPMI està activat, però cap compte actiu té contrasenya IPMI, així que ningú no pot iniciar la sessió. Defineix-ne una a sota.',
        title: 'IPMI',
        warning:
          "L'autenticació IPMI és feble per disseny. Qualsevol que pugui arribar a la placa i conegui un nom d'usuari pot obtenir un hash de la contrasenya IPMI d'aquest usuari i intentar desxifrar-la fora de línia. Feu servir contrasenyes generades, activeu IPMI només en una xarxa de confiança i preferiu Redfish per HTTPS quan l'eina ho admeti.",
        service: 'IPMI per LAN',
        serviceDesc:
          "IPMI 2.0 (RMCP+, ipmitool lanplus) al port UDP 623, per a l'alimentació i l'estat de l'amfitrió. IPMI 1.5 i el conjunt de xifratge 0 es rebutgen. Desactivar-lo tanca totes les sessions IPMI.",
        example: 'Exemple',
        copyFailed: 'La còpia ha fallat. Copieu-ho manualment.',
        ledOn: "Estan disponibles l'estat, l'encesa, l'apagada, soft, cycle i reset.",
        ledOff:
          '"LED d\'alimentació connectat" està desactivat al menú d\'alimentació, de manera que l\'estat d\'alimentació és desconegut. Només funciona "power reset": status, on, off, soft i cycle es rebutgen.',
        accounts: 'Comptes',
        accountsDesc:
          'IPMI inicia la sessió amb els comptes del KVM, cadascun amb la seva pròpia contrasenya IPMI, diferent de la contrasenya web. Els administradors obtenen ADMINISTRATOR. Els usuaris obtenen USER: poden llegir l\'estat d\'alimentació amb "-L USER" però no canviar-lo.',
        passwordSet: 'Contrasenya IPMI definida',
        passwordNotSet: 'Sense contrasenya IPMI: no pot iniciar la sessió per IPMI',
        nameTooLong: 'El nom té més de 16 caràcters, cosa que IPMI no permet',
        accountDisabled: 'El compte està desactivat',
        setPassword: 'Defineix la contrasenya',
        changePassword: 'Canvia la contrasenya',
        remove: 'Elimina',
        removeConfirmTitle: 'Voleu eliminar la contrasenya IPMI de {{user}}?',
        removeConfirmDesc:
          'El compte ja no podrà iniciar la sessió per IPMI, i les seves sessions IPMI es tanquen.',
        passwordTitle: 'Contrasenya IPMI de {{user}}',
        passwordDesc:
          "De 12 a 20 caràcters ASCII imprimibles, diferent de la contrasenya web. IPMI necessita que la placa guardi la contrasenya en una forma que pugui tornar a llegir, així que feu-ne servir una que no s'utilitzi enlloc més. Copieu-la abans de desar: no es tornarà a mostrar.",
        passwordPlaceholder: 'Contrasenya IPMI',
        generate: 'Genera',
        copy: 'Copia',
        save: 'Desa',
        passwordLength: 'Feu servir de 12 a 20 caràcters.',
        passwordChars: 'Feu servir només caràcters ASCII imprimibles.',
        saved: 'Contrasenya IPMI desada',
        failed: "L'operació IPMI ha fallat",
        okBtn: 'Confirma',
        cancelBtn: 'Cancel·la'
      },
      ssh: {
        service: 'Servidor SSH',
        serviceDesc: 'Iniciar sshd ara i a cada arrencada',
        failed: "No s'ha pogut carregar la configuració SSH",
        rootDefault: 'root encara té la contrasenya de fàbrica',
        rootEmpty: 'root no té contrasenya',
        rootWarning:
          'Qualsevol que arribi a la consola o a SSH pot entrar com a root. Definiu una contrasenya a {{account}} > {{password}}: per al propietari del dispositiu també canvia la de root.',
        connection: 'Connexió',
        command: 'Entrar com a root',
        port: 'Port',
        viaVpn: 'Per {{name}}',
        notRunning: "sshd no s'està executant. Activeu el servidor SSH per connectar-vos.",
        hostKeys: "Empremtes de les claus d'amfitrió",
        hostKeysDesc: 'Compareu-les amb el que mostra ssh a la primera connexió.',
        noHostKeys:
          "Encara no hi ha claus d'amfitrió. sshd les crea la primera vegada que arrenca.",
        keys: 'Claus autoritzades',
        keysDesc:
          'Claus públiques que poden entrar com a root. Es desen a la partició de dades, així que les actualitzacions les conserven.',
        noKeys: 'Encara no hi ha claus autoritzades.',
        noComment: 'sense comentari',
        addPlaceholder:
          'Enganxeu una clau pública, per exemple el contingut de ~/.ssh/id_ed25519.pub',
        add: 'Afegeix la clau',
        added: 'Clau afegida',
        removed: 'Clau eliminada',
        deleteConfirm: 'Voleu eliminar aquesta clau?',
        deleteConfirmDesc: 'Ja no podrà entrar. Les sessions obertes continuen obertes.',
        invalidKey: "Això no és una clau pública. Enganxeu una sola línia d'un fitxer .pub.",
        keyOptions: "Aquí no s'accepten claus amb opcions com command= o from=.",
        duplicateKey: 'Aquesta clau ja està autoritzada.',
        lastKey: "L'última clau no es pot eliminar mentre l'accés només amb claus estigui activat.",
        keysOnly: 'Només claus',
        keysOnlyDesc:
          "Desactiva l'accés amb contrasenya i keyboard-interactive. Les sessions obertes continuen obertes.",
        keysOnlyNeedsKey: 'Afegiu primer una clau autoritzada, o ningú no podria entrar.',
        keysOnlyOn: 'Accés amb contrasenya desactivat',
        keysOnlyOff: 'Accés amb contrasenya activat',
        notHonoured:
          "El sshd d'aquesta imatge no llegeix aquest paràmetre, així que l'accés amb contrasenya continua actiu.",
        reloadFailed:
          "Desat, però no s'ha pogut recarregar sshd. S'aplicarà la propera vegada que arrenqui sshd.",
        notApplied:
          'sshd encara accepta contrasenyes. Apagueu i engegueu el servidor SSH per aplicar el paràmetre.',
        changePort: 'Canvia',
        portConfirm: 'Voleu canviar el port SSH a {{port}}?',
        portConfirmDesc:
          'Les sessions SSH actuals continuen obertes. Les connexions noves han de fer servir el port {{port}}. Assegureu-vos que el tallafoc ho permet.',
        portChanged: 'Port SSH canviat a {{port}}',
        portInvalid: "Introduïu un port de l'1 al 65535.",
        portReserved: 'IronKVM ja fa servir aquest port. Trieu-ne un altre.',
        portInUse: "Un altre programa de l'IronKVM ja escolta en aquest port.",
        portNotHonoured:
          "L'sshd d'aquesta imatge no llegeix aquest paràmetre, així que el port no canvia."
      },
      vnc: {
        address: 'Adreça',
        certHint:
          "VeNCrypt X509Plain fa servir el certificat autosignat del dispositiu, així que el client avisa en la primera connexió. Accepta'l, o desa el certificat des de l'adreça HTTPS d'aquesta pàgina i passa'l a TigerVNC amb -X509CA=<fitxer>.",
        title: 'VNC',
        service: 'Servidor VNC',
        serviceDesc:
          "Permet que un client VNC, com ara TigerVNC o Remmina, vegi i controli l'amfitrió. El client ha d'admetre la codificació Tight. Una sessió alhora.",
        credentials:
          'Inicieu la sessió amb un compte KVM. La connexió es xifra amb el certificat TLS de la placa (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'El port TCP on escolta el servidor.',
        maxFps: 'Límit de fotogrames',
        maxFpsDesc: 'El màxim de fotogrames per segon que rep un client.',
        vncAuth: 'Autenticació VNC simple',
        vncAuthDesc:
          "Per a clients sense VeNCrypt. Comprova una contrasenya VNC a part en lloc d'un compte.",
        vncAuthWarning:
          "L'autenticació VNC simple no xifra la connexió. Qualsevol persona al camí de xarxa pot veure la pantalla i les pulsacions de tecles. Feu-la servir només en una xarxa de confiança.",
        password: 'Contrasenya VNC',
        passwordSet: 'Hi ha una contrasenya establerta. Escriviu-ne una de nova per canviar-la.',
        passwordInvalid: 'La contrasenya VNC ha de tenir de 6 a 8 caràcters.',
        save: 'Desa',
        saved: 'Configuració desada',
        state: 'Estat',
        listening: 'Escoltant al port {{port}}',
        notListening: 'No escolta',
        noSession: 'Cap sessió oberta',
        client: 'Client',
        user: 'Usuari',
        method: 'Autenticació',
        methodVencrypt: 'Compte sobre TLS',
        methodVnc: 'Contrasenya VNC',
        since: 'Connectat des de',
        resolution: 'Resolució',
        framesSent: 'Fotogrames enviats',
        lastError: "L'última sessió ha acabat: {{error}}",
        refresh: 'Actualitza',
        disconnect: 'Desconnecta',
        disconnectConfirmTitle: 'Voleu acabar la sessió VNC?',
        disconnectConfirmDesc:
          'El client es desconnecta de seguida, i es deixen anar totes les tecles i botons que mantingui premuts.',
        failed: "L'operació VNC ha fallat",
        okBtn: 'Confirma',
        cancelBtn: 'Cancel·la'
      },
      watchdog: {
        title: 'Watchdog',
        service: "Watchdog de l'amfitrió",
        serviceDesc:
          "Si l'amfitrió hauria d'estar engegat i la seva imatge no canvia, o no hi ha senyal HDMI, durant el temps d'espera, la placa prem reset o apaga i engega l'amfitrió.",
        stillWarning:
          "Un amfitrió amb la pantalla en repòs, o amb la imatge fixa mentre treballa, sembla penjat. Desactiveu el repòs de pantalla a l'amfitrió o indiqueu una adreça de ping.",
        ledHint:
          "\"LED d'alimentació connectat\" està desactivat al menú d'alimentació. El watchdog no veu quan l'amfitrió està apagat, així que el tracta com a sempre engegat.",
        timeout: "Temps d'espera",
        timeoutDesc:
          "Quant de temps pot l'amfitrió no donar senyals de vida abans que el watchdog actuï.",
        action: 'Acció',
        actionDesc:
          "El cicle d'alimentació manté premut el botó d'engegada 5 segons i després el torna a prémer.",
        actionReset: 'Reinicia',
        actionPower: "Cicle d'alimentació",
        cooldown: 'Temps de repòs',
        cooldownDesc: 'El temps mínim entre dues accions.',
        maxPerHour: 'Accions per hora',
        maxPerHourDesc: "El màxim d'accions en una hora.",
        pingHost: 'Adreça de ping',
        pingHostDesc:
          "L'adreça IP de l'amfitrió. Una resposta compta com a senyal de vida. Deixeu-la buida per no fer ping.",
        pingHostInvalid: 'Introduïu una adreça IPv4 o IPv6.',
        minutes: 'min',
        save: 'Desa',
        saved: 'Desat',
        state: 'Detector',
        status: {
          off: 'Desactivat',
          watching: 'Vigilant',
          hostOff: 'Amfitrió apagat',
          captureOff: 'Captura HDMI desactivada',
          cooldown: 'En repòs',
          capped: 'Límit horari assolit',
          acting: 'Actuant'
        },
        signal: 'Senyal HDMI',
        yes: 'Sí',
        no: 'No',
        led: "LED d'alimentació",
        on: 'Encès',
        off: 'Apagat',
        ledNotConnected: 'No connectat',
        ping: 'Ping',
        pingNotSet: 'Sense definir',
        pingReply: 'Respon',
        pingNoReply: 'Sense resposta',
        lastChange: "Últim canvi d'imatge",
        never: 'Mai',
        actsIn: 'Actua en',
        actionsLastHour: "Accions en l'última hora",
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Registre',
        noLog: 'El watchdog encara no ha actuat.',
        refresh: 'Actualitza',
        reasonFrozen: 'La imatge no ha canviat',
        reasonNoSignal: 'Sense senyal HDMI',
        stuckFor: 'sense senyals de vida durant {{duration}}',
        pressFailed: 'La pulsació ha fallat: {{error}}',
        noScreenshot: 'Sense captura',
        failed: 'Ha fallat l’operació del watchdog',
        powerNeedsLed:
          "El cicle d'alimentació necessita «LED d'alimentació connectat» al menú d'alimentació.",
        noLedConfirmTitle: "Activar el watchdog sense el LED d'alimentació?",
        noLedConfirmDesc:
          "La placa no pot veure quan l'amfitrió està apagat, així que el tracta com sempre engegat. Si apagueu l'amfitrió, el watchdog prem el reinici quan passi el temps d'espera. Connecteu el LED d'alimentació per evitar-ho.",
        noLedConfirmOk: 'Activa',
        cancel: 'Cancel·la'
      },
      media: {
        title: 'Mitjans virtuals',
        description:
          "Configuració del diàleg Mitjans de la barra d'eines. Muntar imatges, afegir-les i triar el conjunt de Ventoy es fa al diàleg.",
        ejectFirst:
          'El disc de Ventoy és en una unitat. Expulseu-lo al diàleg Mitjans per desinstal·lar-lo.'
      },
      netboot: {
        title: 'Arrencada per xarxa',
        isoDownload: 'Baixa',
        description:
          "Arrencar l'amfitrió des de la xarxa: iPXE i un menú de les imatges del KVM per l'enllaç de xarxa USB, o netboot.xyz per proxy DHCP a la LAN.",
        addon: "dnsmasq i fitxers d'arrencada",
        addonDesc:
          "Instal·lats a /data: dnsmasq des d'Alpine, iPXE i netboot.xyz des de les seves versions publicades, cadascun comprovat amb la seva suma de verificació.",
        install: 'Instal·la',
        installing: "S'està instal·lant. Pot trigar uns minuts.",
        uninstall: 'Desinstal·la',
        uninstallConfirm:
          "Voleu desactivar l'arrencada per xarxa i eliminar dnsmasq i els fitxers d'arrencada?",
        needsData:
          "L'arrencada per xarxa necessita una imatge d'IronKVM amb la partició /data muntada.",
        usb: "A l'enllaç de xarxa USB",
        usbDesc:
          "Mentre l'enllaç de xarxa USB està actiu, dnsmasq l'atén en lloc d'udhcpd. L'amfitrió rep la seva única adreça sense encaminador ni servidor DNS, iPXE per a la seva arquitectura i un menú de les imatges ISO del KVM.",
        linkOff: "L'enllaç de xarxa USB està desactivat. Activeu-lo a Dispositiu, Xarxa USB.",
        menuUrl: 'Menú',
        leases: "Concessió de l'amfitrió",
        noLeases: 'Cap encara',
        netbootxyzNote:
          "netboot.xyz al menú es carrega des d'internet, on l'enllaç USB no arriba. L'amfitrió necessita internet en un altre port de xarxa.",
        lan: 'Proxy DHCP a la LAN',
        lanDesc:
          "Respon als clients PXE de la LAN amb netboot.xyz, que després carrega el seu menú des d'internet. No assigna mai adreces i no serveix les imatges del KVM.",
        lanWarning:
          "S'ofereix netboot.xyz a tots els clients PXE d'aquesta LAN, no només a l'amfitrió. Activeu-ho només en una xarxa que controleu.",
        lanConfirm: 'Voleu activar el proxy DHCP a la LAN?',
        lanInterface: 'LAN',
        running: 'En execució',
        stopped: 'Aturat',
        images: 'Imatges al menú',
        noImages: "No hi ha imatges ISO al directori d'imatges.",
        boots: 'Arrencades recents',
        noBoots: "L'amfitrió encara no ha descarregat res.",
        log: 'Registre de dnsmasq',
        refresh: 'Actualitza',
        okBtn: 'Confirma',
        cancelBtn: 'Cancel·la',
        failed: "Ha fallat l'operació d'arrencada per xarxa"
      },
      about: {
        title: 'Sobre IronKVM',
        information: 'Informació',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versió aplicació',
        applicationTip: 'Versió de la interfície web de IronKVM',
        image: 'Versió de la imatge',
        imageTip: 'Imatge de targeta IronKVM i la imatge de sistema NanoKVM en què es basa',
        kernel: 'Versió del nucli',
        kernelTip: "Versió del nucli Linux que s'executa ara",
        deviceKey: 'Clau del dispositiu',
        videoMemory: 'Memòria de vídeo',
        videoMemoryTip:
          'Memòria reservada per a la captura de vídeo. No es comparteix amb la resta del sistema.',
        videoMemoryGenerations_one: '{{count}} sessió anterior de IronKVM reté memòria de vídeo',
        videoMemoryGenerations_other:
          '{{count}} sessions anteriors de IronKVM retenen memòria de vídeo',
        videoMemoryReboot: 'Reinicieu per recuperar-la.',
        community: 'Comunitat',
        hostname: 'Nom del dispositiu',
        hostnameUpdated: 'Nom actualitzat. Reinicia per aplicar.',
        ipType: {
          Wired: 'Cablejada',
          Wireless: 'Sense fil',
          Other: 'Altra'
        },
        hostnameInvalid:
          "Feu servir lletres, xifres i guionets, fins a 63 per part separada per punts. Sense guionet a l'inici ni al final d'una part.",
        hostnameFailed: "No s'ha pogut canviar el nom d'amfitrió",
        editHostname: "Edita el nom de l'amfitrió",
        docs: 'Documentació',
        hardware: 'Maquinari',
        hardwareFaq: 'PMF del maquinari',
        disclaimer:
          'IronKVM: firmware comunitari reforçat per al Sipeed NanoKVM. Sense cap vinculació amb Sipeed.',
        basedOn: 'basat en NanoKVM {{version}}'
      },
      preferences: {
        title: 'Preferències'
      },
      performance: {
        title: 'Rendiment'
      },
      appearance: {
        thisBrowser: 'Aquest navegador',
        thisBrowserDesc:
          'Es desa només en aquest navegador. Els altres navegadors conserven la seva.',
        deviceWide: 'Dispositiu',
        deviceWideDesc: "Es desa al dispositiu. S'aplica a tothom qui l'obre.",
        language: 'Idioma',
        languageDesc: "Seleccioneu l'idioma per a la interfície",
        webTitle: 'Títol web',
        webTitleDesc: 'Personalitza el títol de la pàgina',
        menuBar: {
          title: 'Barra de menús',
          mode: 'Mode de visualització',
          modeDesc: 'Mostra la barra de menús a la pantalla',
          modeOff: 'Apagat',
          modeAuto: 'Ocultació automàtica',
          modeAlways: 'Sempre visible',
          keyboardLedStatus: 'Indicadors de bloqueig del teclat',
          keyboardLedStatusDesc:
            'Mostra l’estat de Bloq Num, Bloq Maj i Bloq Despl de l’ordinador remot',
          icons: 'Icones del submenú',
          iconsDesc: 'Mostra les icones del submenú a la barra de menús'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Estat de bloqueig del teclat remot',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Bloq Num',
        numLockShort: 'Num',
        capsLock: 'Bloq Maj',
        capsLockShort: 'Maj',
        scrollLock: 'Bloq Despl',
        scrollLockShort: 'Despl',
        on: 'Activat',
        off: 'Desactivat',
        unknown: 'Desconegut'
      },
      device: {
        title: 'Dispositiu',
        oled: {
          title: 'OLED',
          description: 'Apagar pantalla OLED després de',
          brightness: "Brillantor de l'OLED",
          brightnessDescription: 'Un nivell més baix allarga la vida de la pantalla',
          brightnessLevels: {
            '64': 'Mínima',
            '96': 'Baixa',
            '128': 'Mitjana',
            '160': 'Alta',
            '207': 'Predeterminada',
            '255': 'Màxima'
          },
          0: 'Mai',
          15: '15 s',
          30: '30 s',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 h'
        },
        sections: {
          video: 'Vídeo',
          usb: 'USB',
          frontPanel: 'Panell frontal'
        },
        cpuFreq: {
          title: 'Freqüència de la CPU',
          description: "Defineix la freqüència de la CPU que s'aplica a la propera arrencada",
          tip: "La CPU arrenca a 850 MHz i està especificada per a 1000 MHz. Un valor nou s'aplica a la propera arrencada, no mentre el sistema funciona. 1000 MHz és dins de l'especificació; la temperatura queda molt per sota dels límits amb qualsevol dels dos valors.",
          running: 'En ús: {{mhz}} MHz',
          rebootToApply: 'reinicieu per aplicar-ho',
          rebootConfirm: 'Voleu reiniciar ara per aplicar {{mhz}} MHz?'
        },
        swap: {
          title: 'Swap',
          disable: 'Desactiva',
          description: 'Defineix la mida del fitxer swap',
          tip: 'Pot reduir la vida útil de la targeta SD!',
          active: 'Activa - {{used}} de {{total}}',
          inactive: 'Configurat, però no en ús'
        },
        zram: {
          title: 'Swap comprimida (zram)',
          description: 'Swap a la RAM comprimida, en lloc de a la targeta SD',
          tip: "zram manté la swap fora de la targeta SD, així que no la desgasta. No hi ha swap en disc al darrere: si zram s'omple, el nucli atura un procés en lloc de paginar lentament. El límit de memòria fixa quanta RAM pot ocupar zram.",
          unavailable: 'Els mòduls del nucli no estan instal·lats en aquest dispositiu',
          inactive: "Activada, però el dispositiu no s'ha iniciat",
          active: 'Activa - {{used}} de {{total}}, {{ratio}}x',
          off: 'Desactivada',
          detail: {
            algorithm: 'Algorisme: {{algorithm}}',
            memory: 'Memòria usada: {{used}} de {{limit}}',
            memoryNoLimit: 'Memòria usada: {{used}}, sense límit',
            counters:
              "Pàgines intercanviades: entrada {{in}}, sortida {{out}} (tots els dispositius swap, des de l'arrencada)"
          }
        },
        mouseJiggler: {
          title: 'Mou-ratolí automàtic',
          description: 'Evita que el dispositiu remot entri en repòs',
          disable: 'Desactiva',
          absolute: 'Mode absolut',
          relative: 'Mode relatiu'
        },
        mdns: {
          description: 'Activa descobriment mDNS',
          tip: 'Desactiva-ho si no és necessari'
        },
        hdmi: {
          description: 'Activa la sortida HDMI',
          idleTimeoutTitle: "Temps d'espera d'inactivitat de captura",
          idleTimeoutDescription:
            'Atura la captura HDMI després de no detectar espectadors actius durant',
          minutes: 'min'
        },
        hidOnly: 'Mode només HID',
        hidOnlyDesc: "Deixeu d'emular dispositius virtuals, conservant només el control bàsic HID",
        disk: 'Disc virtual',
        diskDesc: 'Munta un disc U virtual al dispositiu remot',
        network: 'Xarxa virtual',
        networkDesc: 'Munta una targeta de xarxa virtual al dispositiu remot',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Amfitrió:',
          description:
            "Un enllaç de xarxa privat amb l'amfitrió remot pel cable USB. L'amfitrió rep una adreça sense passarel·la ni DNS, de manera que no pot arribar a la vostra LAN a través del IronKVM.",
          mode: 'Protocol',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (per a amfitrions sense NCM)',
          rndis: "RNDIS (ja no s'ofereix)",
          rndisNote: "Aquest enllaç fa servir RNDIS, que ja no s'ofereix. Trieu NCM o ECM.",
          subnet: 'Subxarxa',
          subnetDesc:
            "Una xarxa IPv4 privada, de /24 a /30. El IronKVM pren la primera adreça i l'amfitrió la segona.",
          invalidSubnet: 'Introduïu una subxarxa com ara 172.31.255.0/30.',
          apply: 'Aplica',
          confirm: 'Voleu tornar a connectar el dispositiu USB?',
          reenumerate:
            "En aplicar-ho es reconstrueix la connexió USB. L'amfitrió perd el teclat, el ratolí i el disc virtual durant uns segons."
        },
        audio: 'Altaveu virtual',
        audioDesc:
          "Presenta una targeta de so USB a l'amfitrió remot, perquè el pugueu sentir. L'amfitrió l'ha de seleccionar com a dispositiu de sortida. Canviar-ho reconstrueix la connexió USB.",
        audioNote: "L'àudio està disponible en els dos modes H.264 (WebRTC i Direct), no en MJPEG",
        console: 'Consola sèrie',
        consoleDesc:
          "Presenta un port sèrie USB a l'amfitrió remot, per iniciar la sessió en aquest IronKVM quan la xarxa no és accessible",
        consoleTip:
          "Qualsevol que controli l'amfitrió remot obté una sol·licitud d'inici de sessió d'aquest IronKVM. Configureu una contrasenya segura abans d'activar-ho (Compte - Canvia contrasenya).",
        usbApply: {
          changed: 'Canviat',
          discard: 'Descarta',
          pending: "Els canvis encara no s'han aplicat."
        },
        endpoints: {
          title: 'Ranures USB',
          free: '{{free}} de {{total}} lliures',
          slots: 'Ranures: {{count}}',
          full: 'No hi ha prou ranures USB lliures. Apagueu primer una altra cosa.',
          inactive: "Activat, però no funciona: el controlador USB s'ha quedat sense ranures. Apagueu un altre dispositiu i aquest s'iniciarà de seguida.",
          explain: "El controlador USB té un nombre fix de ranures (endpoints d'entrada), i el teclat i el ratolí sempre n'ocupen algunes. Si hi ha més dispositius activats dels que hi caben, es mantenen el teclat i el ratolí i la resta s'apaguen.",
          error: "No s'ha pogut contactar amb el dispositiu. Torneu-ho a provar.",
          fitTogether: 'Caben junts: {{sets}}'
        },
        reboot: 'Reinicia',
        rebootDesc: 'Segur que vols reiniciar el IronKVM?',
        okBtn: 'Sí',
        cancelBtn: 'No',
        rebootFailed: 'El reinici ha fallat'
      },
      network: {
        title: 'Xarxa',
        wifi: {
          disconnectBtn: 'Desconnecta',
          disconnectWarning:
            'Si accedeixes a IronKVM per aquesta xarxa Wi-Fi, aquesta pàgina perdrà la connexió.',
          disconnected: 'Wi-Fi desconnectada',
          title: 'Wi-Fi',
          description: 'Configura la xarxa Wi-Fi',
          apMode: "El mode AP està activat; connecta't al Wi-Fi escanejant el codi QR",
          connect: "Connecta't a Wi-Fi",
          connectDesc1: 'Introdueix el SSID i la contrasenya de la xarxa',
          connectDesc2: 'Introdueix la contrasenya per connectar-te a aquesta xarxa',
          disconnect: 'Segur que vols desconnectar la xarxa?',
          failed: 'La connexió ha fallat, torna-ho a provar.',
          ssid: 'Nom',
          password: 'Contrasenya',
          joinBtn: 'Connecta',
          confirmBtn: "D'acord",
          cancelBtn: 'Cancel·la'
        },
        tls: {
          description: 'Activa el protocol HTTPS',
          tip: 'Atenció: Usar HTTPS pot augmentar la latència, sobretot amb vídeo MJPEG.',
          restarting: "S'està reiniciant el servidor del dispositiu, triga uns dos minuts...",
          waiting: "S'està esperant que el dispositiu torni a respondre...",
          waitingHttp: "S'està tornant a http. Torneu a carregar aquesta pàgina si no s'obre sola.",
          failed: "No s'ha pogut canviar la configuració HTTPS",
          enableConfirm: 'Activar HTTPS?',
          disableConfirm: 'Desactivar HTTPS?',
          confirmDesc:
            'Això tanca la sessió i reinicia el servidor del dispositiu, cosa que triga uns dos minuts. Després la pàgina obre {{url}}.',
          confirmOk: 'Continua',
          confirmCancel: 'Cancel·la'
        },
        ethernet: {
          title: 'Adreça IP',
          description: "Configureu com IronKVM obté l'adreça a la xarxa amb cable",
          dhcp: 'DHCP',
          manual: 'Manual',
          networkDetails: 'Detalls de la xarxa',
          interface: 'Interfície',
          ipAddress: 'Adreça IP',
          subnetMask: 'Màscara de subxarxa',
          router: 'Encaminador',
          save: 'Aplica',
          invalidAddress: 'Introduïu una adreça IP vàlida',
          invalidMask: 'Introduïu una màscara de subxarxa vàlida, per exemple 255.255.255.0 o 24',
          invalidRouter: "Introduïu una adreça d'encaminador vàlida",
          addressRequired: 'Cal una adreça IP',
          maskRequired: 'Cal una màscara de subxarxa',
          applyTitle: "Voleu canviar l'adreça de IronKVM?",
          applyWarning:
            'Es perdrà la connexió amb aquesta pàgina. IronKVM aplica la nova adreça i espera {{seconds}} segons que hi arribeu. Arribar-hi conserva el canvi. Si no hi arriba res, IronKVM restaura la configuració anterior.',
          applyConfirm: 'Aplica',
          applyCancel: 'Cancel·la',
          applyFailed: "No s'ha pogut aplicar l'adreça",
          trialTitle: 'Esperant confirmació',
          trialDhcp: 'IronKVM està demanant una adreça per DHCP.',
          trialStatic: 'IronKVM ara és a {{address}}.',
          trialInstruction:
            'Obriu IronKVM a la seva nova adreça i inicieu la sessió si us la demana. Arribar-hi conserva el canvi. Si no arriba res a IronKVM en {{seconds}} segons, restaura la configuració anterior.',
          trialOpen: 'Obre la nova adreça',
          trialKeep: 'Conserva aquesta configuració',
          trialKept: "La nova adreça s'ha desat",
          trialKeepFailed: "No s'ha pogut conservar la configuració",
          trialGone: "El canvi ja s'ha restaurat. Torneu-ho a provar.",
          unsaved: 'Canvis no desats'
        },
        dns: {
          title: 'DNS',
          description: 'Configura els servidors DNS per a IronKVM',
          mode: 'Mode',
          dhcp: 'DHCP',
          manual: 'Manual',
          add: 'Afegeix DNS',
          save: 'Desa',
          invalid: 'Introdueix una adreça IP vàlida',
          noDhcp: 'No hi ha cap DNS DHCP disponible actualment',
          saved: 'Configuració DNS desada',
          saveFailed: "No s'ha pogut desar la configuració DNS",
          unsaved: 'Canvis no desats',
          maxServers: 'Es permeten com a màxim {{count}} servidors DNS',
          dnsServers: 'Servidors DNS',
          dhcpServersDescription: "Els servidors DNS s'obtenen automàticament via DHCP",
          manualServersDescription: 'Els servidors DNS es poden editar manualment',
          networkDetails: 'Detalls de xarxa',
          interface: 'Interfície',
          ipAddress: 'Adreça IP',
          subnetMask: 'Màscara de subxarxa',
          router: 'Encaminador',
          none: 'Cap'
        }
      },
      vpn: {
        connect: 'Connecta',
        connectDesc: 'Uneix-te a la xarxa {{name}}. Desactivat desconnecta sense aturar el servei.',
        kvmUrl: 'Adreça del KVM',
        moreTip: 'Més accions',
        restartTip: 'Reinicia',
        stopTip: 'Atura',
        updateTip: 'Actualitza a {{version}}',
        loading: "S'està carregant...",
        okBtn: 'Sí',
        cancelBtn: 'No',
        restart: 'Voleu reiniciar {{name}}?',
        stop: 'Voleu aturar {{name}}?',
        stopDesc:
          "El dimoni s'atura ara. Inicia a l'arrencada és un interruptor a part i es queda com està.",
        update: 'Voleu actualitzar {{name}} a {{version}}?',
        updateDesc: "El dimoni es reinicia si s'està executant. La sessió es conserva.",
        notInstall: '{{name}} no està instal·lat.',
        install: 'Instal·la',
        installing: "S'està instal·lant",
        installFailed: 'La instal·lació ha fallat',
        retry: 'Torna-ho a provar',
        notRunning: "{{name}} no s'està executant. Inicieu-lo per continuar.",
        run: 'Inicia',
        boot: "Inicia a l'arrencada",
        bootDesc: 'Inicia {{name}} quan arrenca el KVM.',
        control: 'Servidor de control',
        connected: 'Connectat',
        disconnected: 'No connectat',
        deviceName: 'Nom del dispositiu',
        deviceIP: 'IP del dispositiu',
        account: 'Compte',
        version: 'Versió',
        uptime: 'Temps actiu',
        peers: 'Iguals',
        noPeers: 'Encara no hi ha iguals.',
        online: 'En línia',
        offline: 'Fora de línia',
        memory: 'Memòria',
        daemonRss: 'Dimoni',
        group: 'Grup de complements',
        high: 'limitat per sobre de {{size}}',
        max: "el nucli l'atura per sobre de {{size}}",
        noGroup: 'Aquesta placa no té grup de memòria per a complements.',
        uninstall: 'Desinstal·la {{name}}',
        uninstallDesc: 'Segur que voleu desinstal·lar {{name}}? La sessió es queda a la placa.',
        blocked:
          "{{other}} s'està executant o s'inicia a l'arrencada. Només pot funcionar una VPN alhora: primer atureu {{other}} i desactiveu-ne l'inici a l'arrencada.",
        swap: {
          title: 'Memòria swap',
          tip: 'Si el dimoni es queda curt de memòria, proveu d\'activar la memòria swap. Es configura a "Configuració > Rendiment".'
        },
        copy: 'Copia',
        copied: 'Enllaç copiat',
        copyFailed: "No s'ha pogut copiar l'enllaç. Seleccioneu-lo i copieu-lo a mà.",
        open: 'Obre',
        checkAgain: 'Comprova de nou',
        notSignedIn:
          "Encara no s'ha iniciat la sessió. Acabeu d'iniciar-la a l'enllaç i torneu-ho a comprovar.",
        checkFailed: "No s'ha pogut comprovar l'estat de la sessió",
        loginWaiting:
          'Aquesta pàgina ho comprova cada pocs segons i continua quan hàgiu iniciat la sessió.',
        uninstallFailed: 'La desinstal·lació ha fallat',
        loginFailed: "No s'ha pogut iniciar la sessió"
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Descarrega el',
        package: "paquet d'instal·lació",
        unzip: 'i descomprimeix-lo',
        notLogin: 'El dispositiu no està vinculat. Inicia sessió per vincular-lo.',
        urlPeriod: 'Aquesta URL és vàlida durant 10 minuts',
        login: 'Inicia sessió',
        logout: 'Tanca sessió',
        logoutDesc: 'Segur que vols tancar sessió?',
        manualIntro: 'O instal·la-ho manualment per SSH:',
        copyBinaries: "Copia tailscale i tailscaled a {{dir}} a l'IronKVM",
        linksFile: 'Al mateix directori, crea un fitxer anomenat links amb aquestes dues línies:',
        rebootRefresh: "Reinicia l'IronKVM i després actualitza aquesta pàgina"
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          "Aquest dispositiu encara no s'ha unit a cap xarxa NetBird. Uniu-vos-hi amb una clau de configuració o inicieu la sessió amb SSO.",
        setupKey: 'Clau de configuració',
        setupKeyPlaceholder: 'Enganxeu una clau de configuració del tauler de NetBird',
        join: 'Uneix-te',
        or: 'o',
        sso: 'Inicia la sessió amb SSO',
        urlPeriod: 'Aquesta URL és vàlida durant 10 minuts',
        logout: 'Dona de baixa',
        logoutDesc:
          "Donar de baixa elimina aquest igual del vostre compte de NetBird i n'esborra aquí la configuració. Per tornar-vos a unir cal una clau de configuració o un inici de sessió amb SSO, i l'igual pot rebre una IP nova. Voleu continuar?",
        joinFailed: "No s'ha pogut unir a la xarxa"
      },
      update: {
        title: 'Comprova actualitzacions',
        queryFailed: 'Error en obtenir la versió',
        updateFailed: 'Error en actualitzar. Torna-ho a intentar.',
        isLatest: 'Ja tens la darrera versió.',
        available: 'Hi ha una actualització disponible. Vols actualitzar ara?',
        updating: 'Actualitzant... espera',
        confirm: 'Confirma',
        cancel: 'Cancel·la',
        preview: 'Versió de prova',
        previewDesc: 'Prova noves funcions abans que ningú',
        previewTip: 'Compte: aquestes versions poden tenir errors o funcions inacabades!',
        customServer: {
          title: 'Servidor d’actualitzacions personalitzat',
          desc: 'Cerca i baixa actualitzacions en línia des d’un servidor especificat',
          invalidUrl:
            'Introduïu un directori de servidor HTTP o HTTPS vàlid, sense paràmetres de consulta, fragments ni latest.json.',
          loadFailed: 'No s’ha pogut carregar la configuració del servidor d’actualitzacions.',
          saveFailed: 'No s’ha pogut desar la configuració del servidor d’actualitzacions.',
          saved: 'S’ha desat la configuració del servidor d’actualitzacions.',
          save: 'Desa',
          confirmTitle: 'Voleu utilitzar un servidor d’actualitzacions personalitzat?',
          confirmDesc:
            'SHA-512 només comprova que el paquet coincideixi amb el manifest proporcionat per aquest servidor. Això no demostra que el paquet sigui una versió oficial de IronKVM. Un servidor defectuós o maliciós pot deixar el dispositiu inutilitzable, provocar la pèrdua de dades o comprometre el sistema.',
          confirm: 'Utilitza’l igualment',
          useSipeed: 'Utilitza el servidor oficial de Sipeed',
          previewDisabled:
            'Les actualitzacions de previsualització no estan disponibles mentre hi hagi activat un servidor d’actualitzacions personalitzat.'
        },
        offline: {
          chooseFile: 'Tria un fitxer',
          installing: 'Pujada completa. Instal·lant...',
          noFile: "No s'ha triat cap fitxer",
          title: 'Actualitzacions fora de línia',
          desc: "Actualització mitjançant el paquet d'instal·lació local",
          upload: 'Puja',
          checksumPlaceholder: 'Suma de verificació SHA-256 (opcional)',
          invalidChecksum:
            'La suma de verificació SHA-256 ha de contenir 64 caràcters hexadecimals.',
          checksumMismatch:
            'La verificació SHA-256 ha fallat. És possible que el paquet estigui malmès.',
          invalidName: 'Format de nom de fitxer no vàlid. Baixeu-lo des de les versions de GitHub.',
          updateFailed: 'Error en actualitzar. Torna-ho a intentar.'
        },
        updateTo: 'Actualitza a {{version}}',
        updateConfirmDesc:
          "El dispositiu instal·la l'actualització i reinicia el servidor. Aquesta pàgina es recarrega quan el servidor torna a respondre.",
        releaseNotes: 'Notes de la versió'
      },
      account: {
        title: 'Compte',
        webAccount: 'Nom del compte web',
        role: 'Rol',
        roles: { admin: 'Administrador', user: 'Usuari' },
        password: 'Contrasenya',
        updateBtn: 'Canvia',
        logoutBtn: 'Tanca sessió',
        logoutDesc: 'Segur que vols tancar sessió?',
        okBtn: 'Sí',
        cancelBtn: 'No',
        users: {
          title: 'Usuaris',
          create: 'Crea un usuari',
          enabled: 'Activat',
          disabled: 'Desactivat',
          deviceOwner: 'Propietari del dispositiu',
          resetPassword: 'Restableix la contrasenya',
          delete: 'Suprimeix',
          deleteConfirm: 'Voleu suprimir aquest usuari i revocar totes les seves sessions?',
          created: "S'ha creat l'usuari",
          deleted: "S'ha suprimit l'usuari",
          passwordUpdated: "S'ha actualitzat la contrasenya",
          loadFailed: "No s'han pogut carregar els usuaris",
          saveFailed: "No s'ha pogut desar l'usuari",
          deleteFailed: "No s'ha pogut suprimir l'usuari"
        }
      },
      apiKeys: {
        mcpNote:
          'Aquestes claus no serveixen per a MCP, que té la seva pròpia clau a la pàgina de MCP.',
        metricsUrl: 'URL de mètriques',
        monitoring: 'Monitoratge',
        monitoringDesc:
          "Prometheus llegeix les mètriques amb una clau API d'aquesta pàgina, enviada com a token Bearer. Qualsevol rol les pot llegir.",
        scrapeConfig: 'Configuració de scrape de Prometheus',
        title: 'Claus API',
        description:
          "Una clau actua com el seu propietari, amb el rol d'aquest usuari. Envieu-la com a Authorization: Bearer <key> per a les mètriques i l'API, o com a X-Auth-Token per a Redfish.",
        name: 'Nom',
        namePlaceholder: 'Per a què és la clau, per exemple prometheus',
        nameRequired: 'Poseu un nom a la clau',
        nameTooLong: 'El nom pot tenir com a màxim 64 caràcters',
        unnamed: '(sense nom)',
        create: 'Crea una clau',
        created: 'Creada',
        owner: 'Propietari',
        empty: 'No hi ha cap clau API',
        newKeyTitle: 'La vostra clau API nova',
        newKeyWarning:
          'Copieu la clau ara. No es desa i no es pot tornar a mostrar. Si la perdeu, revoqueu-la i creeu-ne una altra.',
        copy: 'Copia',
        copied: 'Copiada',
        copyFailed: 'La còpia ha fallat. Copieu-ho manualment.',
        done: 'Fet',
        revoke: 'Revoca',
        revokeConfirmTitle: 'Voleu revocar aquesta clau API?',
        revokeConfirmDesc: 'Tot el que utilitzi "{{name}}" deixarà de funcionar immediatament.',
        revoked: "S'ha revocat la clau API",
        loadFailed: "No s'han pogut carregar les claus API",
        createFailed: "No s'ha pogut crear la clau API",
        revokeFailed: "No s'ha pogut revocar la clau API",
        cancelBtn: 'Cancel·la'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistent',
      empty: 'Obriu el tauler i inicieu una tasca per començar.',
      inputPlaceholder: 'Descriu què vols que faci el PicoClaw',
      newConversation: 'Nova conversa',
      processing: "S'està processant...",
      agent: {
        defaultTitle: 'Assistent general',
        defaultDescription: 'Ajuda general de xat, cerca i espai de treball.',
        kvmTitle: 'Control remot',
        kvmDescription: "Opera l'amfitrió remot mitjançant IronKVM.",
        switched: "Rol d'agent canviat",
        switchFailed: "No s'ha pogut canviar la funció d'agent"
      },
      send: 'Envia',
      cancel: 'Cancel·la',
      status: {
        connecting: "S'està connectant a la passarel·la...",
        connected: 'Sessió de PicoClaw connectada',
        disconnected: 'Sessió de PicoClaw tancada',
        stopped: "S'ha enviat la sol·licitud d'aturada",
        runtimeStarted: "Temps d'execució de PicoClaw iniciat",
        runtimeStartFailed: "No s'ha pogut iniciar el temps d'execució de PicoClaw",
        runtimeStopped: "Temps d'execució de PicoClaw aturat",
        runtimeStopFailed: "No s'ha pogut aturar el temps d'execució de PicoClaw",
        controlSwitchedToMCP: 'El control ha canviat al servei MCP extern'
      },
      connection: {
        runtime: {
          checking: 'Comprovació',
          restoring: 'Restaurant PicoClaw',
          ready: "Temps d'execució a punt",
          stopped: "El temps d'execució s'ha aturat",
          blockedByMCP: 'El control MCP extern està actiu',
          readyBlockedByMCP:
            "El temps d'execució està en marxa, però un MCP extern controla ara l'entrada del dispositiu.",
          readyWithoutControl:
            "El temps d'execució està en marxa. Concedeix a PicoClaw el control del dispositiu abans de tornar a connectar.",
          unavailable: "Temps d'execució no disponible",
          configError: 'Error de configuració'
        },
        transport: {
          connecting: 'En connexió',
          connected: 'Connectat',
          disconnected: 'Desconnectat',
          reconnect: 'Torna a connectar',
          reconnectDescription: 'Torna a connectar a la sessió de PicoClaw en curs.',
          reconnectBlocked:
            'PicoClaw necessita el control del dispositiu abans de tornar a connectar.'
        },
        run: {
          idle: 'Inactiu',
          busy: 'Ocupat'
        }
      },
      message: {
        toolAction: 'Acció',
        observation: 'Observació',
        screenshot: 'Captura de pantalla'
      },
      overlay: {
        locked: "PicoClaw està controlant el dispositiu. L'entrada manual està en pausa."
      },
      control: {
        picoclaw: 'Control del dispositiu: PicoClaw',
        picoclawDescription:
          "PicoClaw pot escriure entrada de teclat i ratolí. L'entrada manual pot quedar en pausa.",
        mcp: 'Control del dispositiu: MCP extern',
        mcpDescription:
          "L'MCP extern pot escriure al dispositiu. PicoClaw no prendrà el control de l'entrada.",
        off: 'Control del dispositiu: desactivat',
        offDescription:
          'La IA no escriurà entrada de teclat ni de ratolí. El control manual continua disponible.',
        transitioning: 'Control del dispositiu: canviant',
        transitioningDescription: "El control del dispositiu s'està sincronitzant. Espera.",
        grant: 'Concedeix control',
        release: 'Allibera',
        releasing: 'Alliberant...',
        switching: 'Canviant...',
        releasingLabel: 'Control del dispositiu: alliberant',
        releasingDescription:
          "S'està retornant el control del dispositiu. PicoClaw ha aturat les escriptures en curs.",
        granted: 'Control de PicoClaw concedit',
        released: 'Control de PicoClaw alliberat',
        grantFailed: "No s'ha pogut concedir el control a PicoClaw",
        releaseFailed: "No s'ha pogut alliberar el control de PicoClaw",
        grantConfirmTitle: 'Vols canviar el control del dispositiu a PicoClaw?',
        grantConfirmDesc: "Les escriptures del dispositiu MCP extern s'interrompran."
      },
      install: {
        install: 'Instal·la PicoClaw',
        installing: 'Instal·lant PicoClaw',
        success: 'PicoClaw instal·lat correctament',
        failed: "No s'ha pogut instal·lar PicoClaw",
        uninstalling: "S'està desinstal·lant el temps d'execució...",
        uninstalled: "El temps d'execució s'ha desinstal·lat correctament.",
        uninstallFailed: 'La desinstal·lació ha fallat.',
        requiredTitle: 'PicoClaw no està instal·lat',
        requiredDescription: "Instal·leu PicoClaw abans d'iniciar el temps d'execució de PicoClaw.",
        progressDescription: "PicoClaw s'està baixant i instal·lant.",
        stages: {
          preparing: 'Preparant',
          downloading: "S'està baixant",
          extracting: 'Extracció',
          verifying: 'Verificant',
          installing: 'Instal·lant',
          installed: 'Instal·lat',
          install_timeout: 'Temps esgotat',
          install_failed: 'Ha fallat'
        }
      },
      model: {
        requiredTitle: 'La configuració del model és necessària',
        requiredDescription: "Configura el model PicoClaw abans d'utilitzar el xat PicoClaw.",
        docsTitle: 'Guia de configuració',
        docsDesc: 'Models i protocols compatibles',
        menuLabel: 'Configura el model',
        modelIdentifier: 'Identificador del model',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Clau API',
        apiKeyPlaceholder: 'Introduïu la clau API del model',
        save: 'Desa',
        saving: 'Desa',
        saved: "S'ha desat la configuració del model",
        saveFailed: "No s'ha pogut desar la configuració del model",
        invalid: "Cal indicar l'identificador del model, l'API Base URL i la clau API"
      },
      uninstall: {
        menuLabel: 'Desinstal·la',
        confirmTitle: 'Desinstal·la PicoClaw',
        confirmContent:
          "Esteu segur que voleu desinstal·lar PicoClaw? Això suprimirà l'executable i tots els fitxers de configuració.",
        confirmOk: 'Desinstal·la',
        confirmCancel: 'Cancel·la'
      },
      history: {
        title: 'Historial',
        loading: 'Carregant sessions...',
        emptyTitle: 'Encara no hi ha historial',
        emptyDescription: 'Les sessions anteriors de PicoClaw apareixeran aquí.',
        loadFailed: "No s'ha pogut carregar l'historial de sessions",
        deleteFailed: "No s'ha pogut suprimir la sessió",
        deleteConfirmTitle: 'Suprimeix la sessió',
        deleteConfirmContent: 'Esteu segur que voleu suprimir "{{title}}"?',
        deleteConfirmOk: 'Esborra',
        deleteConfirmCancel: 'Cancel·la',
        messageCount_one: '{{count}} missatge',
        messageCount_other: '{{count}} missatges',
        messageCount: '{{count}} missatges'
      },
      config: {
        startRuntime: 'Inici PicoClaw',
        stopRuntime: 'Atura PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Voleu canviar el control a PicoClaw?',
        enableConfirmDesc: 'En iniciar PicoClaw es desactivarà el servei MCP extern.',
        enableConfirmOk: 'Inicia PicoClaw',
        enableConfirmCancel: 'Cancel·la',
        title: 'Inici PicoClaw',
        description: "Inicieu el temps d'execució per començar a utilitzar l'assistent PicoClaw.",
        switchFromMCP: 'Canvia a PicoClaw i inicia',
        takeoverAndStart: 'Pren el control i inicia'
      }
    },
    error: {
      title: 'Hi ha hagut un error',
      refresh: 'Actualitza',
      panel: 'Aquesta part de la pàgina ha deixat de funcionar',
      retry: 'Torna-ho a provar'
    },
    fullscreen: {
      toggle: 'Pantalla completa'
    },
    input: {
      disconnected: 'El teclat i el ratolí no estan connectats',
      disconnectedTls:
        'El navegador ha rebutjat la connexió segura que transporta el teclat i el ratolí, cosa que fa sense preguntar. El certificat que ha generat aquest dispositiu encara no és de confiança. Obriu aquesta adreça en una pestanya nova, accepteu el certificat i torneu a carregar. Instal·lar el certificat és la solució fiable.',
      disconnectedNever:
        "No s'ha pogut obrir la connexió que transporta el teclat i el ratolí. La resta de la pàgina funciona perquè no la fa servir. Comproveu que res entre vosaltres i el dispositiu la bloquegi.",
      disconnectedDropped:
        "La connexió que transporta el teclat i el ratolí s'ha perdut i no s'ha recuperat. Es torna a connectar sola després d'un reinici; si continua així, torneu a carregar la pàgina.",
      hidDisabled: "L'HID està desactivat en aquest dispositiu (/boot/disable_hid).",
      keyFailed: "No s'ha pogut enviar la tecla."
    },
    speaker: { title: 'Altaveu', unmute: 'Activa el so', mute: 'Silencia' },
    upstream: {
      check: 'Cerca actualitzacions',
      updateTo: 'Actualitza a {{version}}',
      confirm: 'Vols actualitzar {{name}} a {{version}}?',
      confirmDesc:
        'La nova versió es baixa de GitHub i es comprova amb les sumes de verificació que publica. Si alguna cosa falla, es manté la versió actual.',
      ok: 'Actualitza',
      upToDate: 'Actualitzat',
      builtIn: 'integrada',
      checkFailed: "No s'han pogut cercar actualitzacions: {{error}}",
      unverifiable: "La versió {{version}} no s'ofereix: {{reason}}",
      inUse: 'Ara no es pot actualitzar: {{reason}}',
      running: 'Actualitzant a {{version}}...',
      done: '{{name}} actualitzat a {{version}}',
      failed: "L'última actualització ha fallat: {{error}}"
    },
    menu: {
      mediaAdd: 'Afegeix una imatge',
      mediaMoreOptions: 'Més opcions',
      mediaSettings: 'Configuració de mitjans',
      collapse: 'Amaga menú',
      expand: 'Mostra menú',
      more: 'Més',
      media: 'Mitjans',
      tools: 'Eines',
      text: 'Text',
      advanced: 'Avançat',
      mediaMounted: 'Muntat',
      mediaLibrary: 'Biblioteca',
      textToHost: "Cap a l'amfitrió",
      textFromHost: "Des de l'amfitrió"
    },
    ion: {
      checking: "S'està comprovant la memòria de vídeo abans d'iniciar la transmissió...",
      warn: "Queda poca memòria de vídeo. Un sol reinici del servidor l'esgotaria. Reinicieu quan us vagi bé.",
      criticalTitle: 'No hi ha prou memòria de vídeo per iniciar la transmissió',
      criticalBody:
        "Iniciar el vídeo esgotaria la memòria reservada i aturaria el servidor. Totes les altres funcions continuen funcionant, inclosos el control d'alimentació i el reinici. Només un reinici del IronKVM recupera aquesta memòria.",
      criticalContinue: 'Inicia el vídeo igualment',
      criticalReboot: 'Reinicia el IronKVM',
      criticalRebooting: "S'està reiniciant..."
    }
  }
};

export default ca;
