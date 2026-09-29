const de = {
  translation: {
    feedback: {
      enabled: '{{name}} aktiviert',
      disabled: '{{name}} deaktiviert',
      failed: 'Die Anfrage ist fehlgeschlagen. Bitte erneut versuchen.',
      network:
        'Das Gerät ist nicht erreichbar. Prüfen Sie die Verbindung und versuchen Sie es erneut.',
      saved: 'Gespeichert',
      timeout: 'Das Gerät hat zu lange nicht geantwortet. Bitte erneut versuchen.'
    },
    common: {
      copy: 'Kopieren',
      copied: 'Kopiert',
      copyFailed: 'Kopieren fehlgeschlagen. Markieren Sie den Text und kopieren Sie ihn manuell.',
      notUpdating: 'Keine Aktualisierung: Die letzte Abfrage ist fehlgeschlagen.',
      off: 'Aus',
      running: 'Läuft',
      save: 'Speichern',
      cancel: 'Abbrechen',
      delete: 'Löschen',
      remove: 'Entfernen'
    },
    head: {
      desktop: 'Entfernter Desktop',
      login: 'Anmelden',
      changePassword: 'Passwort ändern',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Passwort geändert. Melden Sie sich mit dem neuen Passwort an.',
      cookieRejected:
        'Der Browser hat das Speichern der Sitzung verweigert. Ein Cookie aus einer früheren HTTPS-Sitzung kann über unverschlüsseltes http nicht ersetzt werden. Löschen Sie die Cookies für diese Adresse oder öffnen Sie ein privates Fenster und melden Sie sich erneut an.',
      login: 'Anmelden',
      placeholderUsername: 'Benutzername',
      placeholderPassword: 'Passwort',
      placeholderCurrentPassword: 'Aktuelles Passwort',
      placeholderPassword2: 'Bitte Passwort erneut eingeben',
      noEmptyUsername: 'Benutzername benötigt',
      noEmptyPassword: 'Passwort benötigt',
      passwordLength: 'Das Passwort muss zwischen 8 und 72 Zeichen lang sein',
      noAccount:
        'Abfragen der Benutzerdaten fehlgeschlagen, bitte die Seite neu laden oder Passwort zurücksetzen',
      invalidUser: 'Falscher Benutzername oder falsches Passwort',
      locked: 'Zu viele Anmeldungen, bitte versuchen Sie es später noch einmal',
      globalLocked: 'System wird geschützt, bitte versuchen Sie es später erneut',
      error: 'Unerwarteter Fehler',
      invalidCurrentPassword: 'Das aktuelle Passwort ist falsch',
      changePassword: 'Passwort ändern',
      changePasswordDesc: 'Für die Sicherheit Ihres Geräts ändern Sie bitte das Passwort!',
      differentPassword: 'Passwörter stimmen nicht überein',
      illegalUsername: 'Benutzername enthält ungültige Zeichen',
      illegalPassword: 'Passwort enthält ungültige Zeichen',
      forgetPassword: 'Passwort vergessen',
      ok: 'Ok',
      cancel: 'Abbrechen',
      loginButtonText: 'Anmelden',
      tips: {
        reset1:
          'Um das Passwort zurückzusetzen, drücken und halten Sie den BOOT Knopf auf dem IronKVM für 10 Sekunden.',
        reset3: 'Web Standard-Account:',
        reset4: 'SSH Standard-Account:',
        change1: 'Bitte beachten Sie, dass diese Aktion folgende Passwörter ändert:',
        change2: 'Web Anmelde-Passwort',
        change3: 'System root Passwort (SSH Anmelde-Passwort)',
        change4:
          'Um die Passwörter zurückzusetzen, drücken und halten Sie den BOOT Knopf auf dem IronKVM.',
        resetDocs: 'Die genauen Schritte stehen in der Hardware-Dokumentation:',
        hardwareDocs: 'Sipeed-NanoKVM-Wiki'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Wi-Fi Konfiguration für IronKVM',
      success:
        'Bitte überprüfen Sie den Netzwerk-Status des IronKVM und greifen Sie über die neue IP Adresse darauf zu.',
      failed: 'Aktion fehlgeschlagen, bitte erneut versuchen.',
      invalidMode:
        'Der aktuelle Modus unterstützt keine Netzwerkeinrichtung. Bitte gehen Sie zu Ihrem Gerät und aktivieren Sie den Wi-Fi-Konfigurationsmodus.',
      confirmBtn: 'Ok',
      finishBtn: 'Fertig',
      ap: {
        authTitle: 'Authentifizierung erforderlich',
        authDescription: 'Bitte geben Sie das Passwort AP ein, um fortzufahren',
        authFailed: 'Ungültiges AP Passwort',
        passPlaceholder: 'AP Passwort',
        verifyBtn: 'Überprüfen'
      },
      ssidRequired: 'Netzwerknamen eingeben, bis zu 32 Zeichen',
      passwordLength: 'Das Passwort hat 8 bis 63 Zeichen. Für ein offenes Netzwerk leer lassen.',
      passwordOptional: 'Passwort (leer für ein offenes Netzwerk)',
      lost: 'Das Board antwortet nicht mehr. Es hat sich vielleicht mit dem Netzwerk verbunden und seinen Einrichtungs-Hotspot geschlossen. Erscheint der Hotspot wieder, ist die Verbindung fehlgeschlagen: erneut mit ihm verbinden und noch einmal versuchen.',
      done: 'Einrichtung abgeschlossen. Dieses Gerät wieder mit dem üblichen Netzwerk verbinden und das Board unter seiner neuen Adresse öffnen.'
    },
    screen: {
      viewOnly: 'Nur ansehen',
      viewOnlyTip:
        'Dieser Tab sendet keine Tastatur- und Mauseingaben mehr an den Host. Skripte, der Maus-Jiggler und andere Zuschauer sind nicht betroffen.',
      viewOnlyOff: 'Nur ansehen ausschalten',
      viewOnlyBlocked: 'Nur ansehen ist aktiv, daher wurde nichts an den Host gesendet',
      pauseHidden: 'Pausieren, wenn Tab verborgen',
      pauseHiddenTip:
        'Stoppt Video und Ton einige Sekunden nachdem dieser Tab verborgen wurde und startet sie wieder, wenn Sie zurückkehren.',
      screenshot: 'Bildschirmfoto',
      screenshotTip: 'Speichert den Bildschirm des Hosts als PNG-Datei in voller Aufnahmegröße.',
      screenshotFailed: 'Bildschirmfoto fehlgeschlagen',
      stream: {
        ok: 'Bild OK',
        noSignal: 'kein Signal',
        failed: 'Stream fehlgeschlagen'
      },
      codecNoWebrtcHevc: 'Dieser Browser kann H.265 nicht über WebRTC empfangen',
      codecNoHevc: 'Dieser Browser kann H.265 nicht dekodieren',
      codecNote:
        'Das Board hat einen Encoder, daher ändert dies den Stream für alle Zuschauer. Für eine laufende WebRTC-Sitzung neu verbinden.',
      codec: 'Codec',
      updateFailed: 'Die Einstellung wurde nicht übernommen',
      scale: 'Skala',
      title: 'Bildschirm',
      video: 'Video Modus',
      videoDirectTips:
        'Aktivieren Sie HTTPS unter „Einstellungen > Gerät“, um diesen Modus zu verwenden',
      resolution: 'Auflösung',
      ocr: {
        title: 'Text lesen (OCR)',
        tips: 'Der Text wird in diesem Browser erkannt. Sie können ihn vor dem Kopieren korrigieren.',
        hint: 'Ziehen Sie über den Text, der gelesen werden soll. Mit Esc brechen Sie ab.',
        noPicture:
          'Warten Sie auf das Video und ziehen Sie dann über den Text, der gelesen werden soll.',
        cancel: 'Abbrechen',
        language: 'Sprache',
        languages: {
          eng: 'Englisch'
        },
        preview: 'Ausgewählter Bereich',
        capturing: 'Bildschirm wird erfasst...',
        loading: 'Texterkennung wird geladen...',
        recognizing: 'Text wird gelesen...',
        noText: 'Im ausgewählten Bereich wurde kein Text gefunden.',
        copy: 'Kopieren',
        copied: 'In die Zwischenablage kopiert',
        copyFailed: 'Kopieren in die Zwischenablage fehlgeschlagen',
        selectAgain: 'Erneut auswählen',
        unsupported:
          'Dieser Browser kann keine Texterkennung ausführen. Sie benötigt WebAssembly SIMD, das aktuelle Browser unterstützen.',
        captureFailed: 'Der Bildschirm konnte nicht erfasst werden.',
        outside: 'Der ausgewählte Bereich liegt außerhalb des Bildes.',
        recognizeFailed: 'Die Texterkennung ist fehlgeschlagen.'
      },
      controlRegion: {
        title: 'Mauskalibrierung',
        description:
          'Verwenden Sie diese Einstellung, wenn das gesteuerte Gerät eine andere Auflösung als 16:9 verwendet und der Mauszeiger horizontal oder vertikal versetzt ist.',
        off: 'Aus',
        auto: 'Automatisch',
        autoWarning:
          'Die Kalibrierung kann fehlschlagen, wenn die Benutzeranwendung einen vollständig schwarzen Hintergrund hat.',
        manual: 'Manuell',
        selectedResolution: 'Auflösung des ausgewählten Bereichs',
        unused: 'Nicht verwendet',
        originalResolution: 'Originalauflösung',
        selectResolution: 'Originalauflösung auswählen',
        addResolution: 'Benutzerdefinierte Auflösung hinzufügen',
        add: 'Hinzufügen',
        duplicateResolution: 'Diese Auflösung ist bereits vorhanden.',
        width: 'Breite',
        height: 'Höhe',
        apply: 'Berechnen und anwenden',
        invalidResolution:
          'Geben Sie eine gültige Originalauflösung ein, sobald das Video bereit ist.',
        select: 'Bereich auswählen',
        clear: 'Automatische Erkennung wiederherstellen',
        saveFailed: 'Der Eingabebereich konnte nicht gespeichert werden.',
        tooSmall: 'Der ausgewählte Bereich ist zu klein.',
        previewUnavailable: 'Vorschau nicht verfügbar',
        clearConfirm: 'Automatische Erkennung schwarzer Ränder wiederherstellen?',
        dragHint: 'Ziehen Sie, um den Remote-Desktop-Bereich auszuwählen',
        finish: 'Fertig',
        confirm: 'Bestätigen',
        cancel: 'Abbrechen'
      },
      auto: 'Automatisch',
      autoTips:
        'Bildverzerrungen oder ein versetzter Mauszeiger können bei bestimmten Auflösungen auftreten. Versuchen Sie, die Auflösung des entfernten Hosts anzupassen oder den automatischen Modus zu deaktivieren.',
      fps: 'FPS',
      customizeFps: 'Anpassen',
      quality: 'Qualität',
      qualityLossless: 'Beste',
      qualityHigh: 'Hoch',
      qualityMedium: 'Mittel',
      qualityLow: 'Niedrig',
      frameDetect: 'Bilderkennung',
      frameDetectTip:
        'Berechnet den Unterschied zwischen den Einzelbildern. Beendet die Liveübertragung des Videostreams wenn keine Änderungen auf dem Bildschirm des Hosts festgestellt werden kann.',
      resetHdmi: 'HDMI zurücksetzen',
      mixedH264: {
        title: 'H.264-Streamkonflikt',
        description:
          'H.264 Direct und H.264 WebRTC werden gleichzeitig verwendet. Dies kann zu Bildschirm-Tearing oder beschädigtem Video führen. Bitte verwenden Sie nur einen H.264-Modus.'
      },
      webrtcConnectionFailed: {
        title: 'WebRTC-Verbindung fehlgeschlagen',
        description: 'Überprüfen Sie die Netzwerkverbindung oder wechseln Sie den Videomodus.'
      },
      captureStatus: {
        hdmiError: 'HDMI-Bildschirmfehler',
        unsupportedResolution: 'Die aktuelle Auflösung wird nicht unterstützt',
        retrieving: 'Bildschirm wird abgerufen...',
        changingResolution: 'Auflösung wird gewechselt...',
        updateFailed: 'Der Bildschirm kann derzeit nicht aktualisiert werden',
        videoError: 'Fehler bei der Videoanzeige',
        noHdmi: 'Kein HDMI-Signal erkannt',
        unavailable: 'Der Bildschirm kann derzeit nicht angezeigt werden'
      },
      directConnectionFailed: 'Verbindung zum Videostream fehlgeschlagen'
    },
    keyboard: {
      close: 'Schließen',
      title: 'Tastatur',
      paste: 'Einfügen',
      tips: 'Tippt den Text auf dem Host als Tastendrücke. Wählen Sie das Tastaturlayout des Hosts.',
      placeholder: 'Bitte eingeben',
      submit: 'Senden',
      virtual: 'Tastatur',
      readClipboard: 'Aus der Zwischenablage lesen',
      clipboardPermissionDenied:
        'Berechtigung für die Zwischenablage verweigert. Bitte erlauben Sie den Zugriff auf die Zwischenablage in Ihrem Browser.',
      clipboardReadError: 'Zwischenablage konnte nicht gelesen werden',
      mediaKeys: {
        title: 'Medientasten',
        mute: 'Stumm',
        volumeDown: 'Leiser',
        volumeUp: 'Lauter',
        previous: 'Vorheriger Titel',
        playPause: 'Wiedergabe oder Pause',
        next: 'Nächster Titel',
        stop: 'Stopp'
      },
      pasting: {
        layout: 'Tastaturlayout des Hosts',
        layouts: {
          us: 'Englisch (USA)',
          uk: 'Englisch (Vereinigtes Königreich)',
          de: 'Deutsch',
          fr: 'Französisch',
          es: 'Spanisch',
          it: 'Italienisch',
          ptBr: 'Portugiesisch (Brasilien)',
          se: 'Schwedisch / Finnisch',
          ru: 'Russisch',
          ja: 'Japanisch',
          ko: 'Koreanisch'
        },
        speed: 'Tippgeschwindigkeit',
        speeds: {
          fast: 'Schnell',
          normal: 'Normal',
          slow: 'Langsam'
        },
        estimate: 'Tippdauer: etwa {{duration}}',
        untypeable: 'Zeichen, die dieses Layout nicht tippen kann: {{count}}',
        untypeableAt: 'Zeile {{line}}, Spalte {{column}}',
        skipUntypeable: 'Den Rest tippen',
        shortcut: '{{shortcut}} tippt die Zwischenablage sofort auf dem Host.',
        clipboardUnavailable:
          'Der Browser lässt eine Seite die Zwischenablage nur über HTTPS lesen. Fügen Sie den Text mit Strg+V in das Feld ein.',
        clipboardEmpty: 'Die Zwischenablage enthält keinen Text.',
        tooLong: 'Der Text ist zu lang. Die Grenze liegt bei {{max}} Zeichen.',
        inProgress: 'Es wird bereits ein Text getippt.',
        typing: 'Tippt auf dem Host',
        done: 'Text getippt',
        canceled: 'Einfügen abgebrochen',
        failed: 'Einfügen fehlgeschlagen',
        cancel: 'Abbrechen',
        controlBusy: 'Eine andere Steuerung verwendet die Tastatur.',
        hidError: 'Die Tastendrücke konnten nicht an den Host gesendet werden.'
      },
      shortcut: {
        sendFailed: 'Nicht gesendet: Die Eingabeverbindung ist unterbrochen',
        title: 'Verknüpfungen',
        custom: 'Benutzerdefiniert',
        capture: 'Klicken Sie hier, um die Verknüpfung zu erfassen',
        clear: 'Klar',
        save: 'Speichern',
        captureTips:
          'Das Erfassen systemweiter Tasten (z. B. der Windows-Taste) erfordert die Vollbildberechtigung.',
        enterFullScreen: 'Vollbildmodus umschalten.'
      },
      leaderKey: {
        saveFailed: 'Leader-Taste konnte nicht gespeichert werden',
        title: 'Leader-Taste',
        desc: 'Browserbeschränkungen umgehen und Systemverknüpfungen direkt an den Remote-Host senden.',
        howToUse: 'Verwendung',
        simultaneous: {
          title: 'Simultanmodus',
          desc1: 'Halten Sie die Leader-Taste gedrückt und drücken Sie dann die Tastenkombination.',
          desc2: 'Intuitiv, kann jedoch zu Konflikten mit Systemverknüpfungen führen.'
        },
        sequential: {
          title: 'Sequenzieller Modus',
          desc1:
            'Drücken Sie die Leader-Taste → drücken Sie die Tastenkombination nacheinander → drücken Sie erneut die Leader-Taste.',
          desc2: 'Erfordert mehr Schritte, vermeidet jedoch vollständig Systemkonflikte.'
        },
        enable: 'Leader-Taste aktivieren',
        tip: 'Wenn diese Taste als Leader-Taste zugewiesen wird, dient sie ausschließlich als Auslöser für Tastenkombinationen und verliert ihr Standardverhalten.',
        placeholder: 'Bitte drücken Sie die Leader-Taste',
        shiftRight: 'Rechts Shift',
        ctrlRight: 'Rechts Ctrl',
        metaRight: 'Rechts Win',
        submit: 'Senden',
        recorder: {
          rec: 'REC',
          activate: 'Tasten aktivieren',
          input: 'Bitte drücken Sie die Tastenkombination...'
        }
      }
    },
    mouse: {
      jiggler: 'Maus-Jiggler',
      title: 'Maus',
      cursor: 'Cursor',
      default: 'Standard Cursor',
      pointer: 'Zeiger Cursor',
      cell: 'Zellen Cursor',
      text: 'Text Cursor',
      grab: 'Greif Cursor',
      hide: 'Versteckter Cursor',
      mode: 'Maus Modus',
      absolute: 'Absoluter Modus',
      relative: 'Relativer Modus',
      absoluteShort: 'Absolut',
      relativeShort: 'Relativ',
      touch: 'Touch-Modus',
      touchShort: 'Touch',
      absoluteStalled: 'Das Zielgerät ignoriert die absolute Maus',
      absoluteStalledDesc:
        'Das Zielgerät nimmt keine absoluten Mausberichte mehr an, daher gehen Zeigerbewegungen verloren. Die Tastatur ist nicht betroffen. Ein Wiederherstellen der USB-Verbindung behebt das oft; der relative Modus nutzt einen anderen Endpunkt.',
      useRelative: 'Zum relativen Modus wechseln',
      direction: 'Scrollrichtung',
      scrollUp: 'Wie auf diesem Computer',
      scrollDown: 'Umgekehrt (natürliches Scrollen)',
      speed: 'Scrollgeschwindigkeit',
      fast: 'Schnell',
      slow: 'Langsam',
      requestPointer:
        'Relativer Modus aktiv. Klicken Sie auf den Desktop um den Mauszeiger zu sehen.',
      resetHid: 'HID zurücksetzen',
      hidOnly: {
        switchFailed:
          'Der Modus konnte nicht gewechselt werden. Verbindung prüfen und erneut versuchen.',
        title: 'HID-Only-Modus',
        desc: 'Wenn Ihre Maus und Tastatur nicht mehr reagieren und das Zurücksetzen der HID-Verbindung nicht hilft, könnte es sich um ein Kompatibilitätsproblem zwischen dem IronKVM und dem Gerät handeln. Versuchen Sie, den HID-Only Modus zu aktivieren, um die Kompatibilität zu verbessern.',
        tip1: 'Die Aktivierung des HID-Only Modus entfernt das virtuelle U-Laufwerk und das virtuelle Netzwerk.',
        tip2: 'Im HID-Only Modus ist das Einbinden von Systemabbilder deaktiviert.',
        rebuild:
          'Beim Wechsel des Modus wird die USB-Verbindung neu aufgebaut. IronKVM startet nicht neu',
        enable: 'HID-Only Modus aktivieren',
        disable: 'HID-Only Modus deaktivieren'
      },
      resetHidDone: 'USB-HID zurückgesetzt',
      resetHidFailed: 'USB-HID konnte nicht zurückgesetzt werden'
    },
    image: {
      driveLoaded: 'Image eingelegt',
      driveWarning: 'Warnungen beachten',
      warning: {
        missing:
          'Die Image-Datei wurde gelöscht. Der Host liest die alte Kopie weiter, bis Sie sie auswerfen.',
        writable: 'Lesen und Schreiben: Der Host kann dieses Image verändern.',
        tooBigForCd:
          'Zu groß für das CD-Laufwerk ({{size}}, Grenze {{max}}). Verwenden Sie das Festplattenlaufwerk.',
        tooSmallForCd:
          'Zu klein für das CD-Laufwerk ({{size}}). Verwenden Sie das Festplattenlaufwerk.',
        empty: 'Die Datei ist leer, vermutlich nach einem fehlgeschlagenen Upload oder Download.'
      },
      delete: 'Löschen',
      inUse: 'In Verwendung. Vor dem Löschen auswerfen.',
      retry: 'Erneut versuchen',
      loadFailed: 'Die Image-Liste konnte nicht geladen werden',
      readOnlyLocked:
        'Zum Ändern die Disk auswerfen. Die Einstellung gilt beim Einlegen eines Images.',
      title: 'Bilder',
      loading: 'Lädt...',
      empty: 'Nichts gefunden',
      mountMode: 'Mount-Modus',
      mountFailed: 'Einbinden fehlgeschlagen',
      mountDesc:
        'In einigen Systemen ist es notwendig, die virtuelle Festplatte auf dem entfernten Host auszuwerfen, bevor das Image eingebunden werden kann.',
      unmountFailed: 'Das Aufheben der Bereitstellung ist fehlgeschlagen',
      unmountDesc:
        'Auf einigen Systemen müssen Sie das Image manuell vom Remote-Host auswerfen, bevor Sie die Bereitstellung aufheben.',
      refresh: 'Bilder aktualisieren',
      disk: 'Festplatte',
      cdrom: 'CD',
      driveEmpty: 'Leer',
      eject: 'Auswerfen',
      readOnly: 'Schreibgeschützt',
      readOnlyTip: 'Gilt für das nächste Image, das in das Laufwerk eingelegt wird.',
      noDrives:
        'Keine virtuellen Laufwerke. Aktivieren Sie die virtuelle Festplatte in den Einstellungen.',
      insertFailed: 'Einlegen fehlgeschlagen',
      ejectFailed: 'Auswerfen fehlgeschlagen',
      insertInto: 'In {{drive}} einlegen. Zum Ändern klicken.',
      loadedIn: 'Im Laufwerk {{drive}}',
      attention: 'Achtung',
      deleteConfirm: 'Sind Sie sicher, dass Sie dieses Bild löschen möchten?',
      okBtn: 'Ja',
      cancelBtn: 'Nein',
      deleteFailed: 'Löschen fehlgeschlagen',
      ventoy: {
        statusNoKernel: 'Von dieser Firmware nicht unterstützt',
        statusNotInstalled: 'Nicht installiert',
        statusReady: 'Bereit',
        statusSelected: 'Ausgewählte Images: {{count}}',
        statusInDrive: 'Im Laufwerk, {{size}}',
        noKernel:
          'Der Kernel dieser Firmware unterstützt kein Device-Mapper, daher kann Ventoy erst nach der Installation eines Images mit dieser Unterstützung genutzt werden.',
        installDesc: 'Den Host von mehreren Images auf einer Disk booten, ohne sie zu kopieren.',
        install: 'Installieren',
        installing: 'Ventoy wird heruntergeladen, etwa 20 MB. Das kann einige Minuten dauern.',
        needsData: 'Ventoy benötigt ein IronKVM-Image mit eingehängter /data-Partition.',
        uninstall: 'Deinstallieren',
        uninstallConfirm: 'Ventoy-Dateien entfernen?',
        noImages: 'Keine Images für die Ventoy-Disk.',
        onDisk: 'Auf der Ventoy-Disk',
        missing: 'Fehlt: {{file}}',
        remove: 'Von der Ventoy-Disk entfernen',
        setHint:
          'Die Auswahl der Images lässt sich nur ändern, solange die Ventoy-Disk in keinem Laufwerk ist.',
        useAsDisk: 'Als virtuelle Disk verwenden',
        failed: 'Ventoy-Anfrage fehlgeschlagen',
        secureBoot:
          'Bei aktivem Secure Boot muss der Host den Schlüssel von Ventoy einmalig in MokManager registrieren. Die Schlüsseldatei ENROLL_THIS_KEY_IN_MOKMANAGER.cer liegt auf der Partition VTOYEFI.',
        readOnly:
          'Der Host sieht die Disk schreibgeschützt, daher funktionieren Ventoy-Persistenz und ventoy.json auf dem Laufwerk nicht.'
      },
      tips: {
        title: 'So laden Sie Dateien hoch',
        usb1: 'Verbinden Sie den IronKVM über USB mit Ihrem Computer.',
        usb2: 'Stellen Sie sicher, dass die virtuelle Festplatte eingebunden ist (Einstellungen – Virtuelle Festplatte).',
        usb3: 'Öffnen Sie die virtuelle Festplatte auf Ihrem Computer und kopieren Sie die Image-Datei in das Stammverzeichnis der virtuellen Festplatte.',
        scp1: 'Stellen Sie sicher, dass sich der IronKVM und Ihr Computer im selben lokalen Netzwerk befinden.',
        scp2: 'Öffnen Sie ein Terminal auf Ihrem Computer und verwenden Sie den SCP-Befehl, um die Image-Datei in das Verzeichnis /data auf dem IronKVM hochzuladen.',
        scp3: 'Beispiel: scp your-image-path root@your-ironkvm-ip:/data',
        tfCard: 'TF-Karte',
        tf1: 'Diese Methode wird unter Linux-Systemen unterstützt.',
        tf2: 'Entnehmen Sie die TF-Karte aus dem IronKVM (bei der FULL-Version muss zuvor das Gehäuse geöffnet werden).',
        tf3: 'Stecken Sie die TF-Karte in einen Kartenleser und verbinden Sie diesen mit Ihrem Computer.',
        tf4: 'Kopieren Sie die Image-Datei in das Verzeichnis /data auf der TF-Karte.',
        tf5: 'Setzen Sie die TF-Karte wieder in den IronKVM ein.'
      }
    },
    script: {
      title: 'Skripte',
      upload: 'Hochladen',
      run: 'Ausführen',
      runBackground: 'Im Hintergrund ausführen',
      runFailed: '',
      attention: 'Achtung',
      delDesc: 'Möchten Sie diese Datei wirklich löschen?',
      confirm: 'Ja',
      cancel: 'Nein',
      delete: 'Löschen',
      close: 'Schliessen',
      empty:
        'Noch keine Skripte. Eine .sh- oder .py-Datei hochladen, um sie auf dem Board auszuführen.',
      loadFailed: 'Skripte konnten nicht geladen werden',
      uploaded: 'Skript hochgeladen',
      uploadFailed: 'Skript konnte nicht hochgeladen werden',
      started: 'Skript im Hintergrund gestartet',
      deleteFailed: 'Skript konnte nicht gelöscht werden',
      waitLimit: 'Warte bis zu {{minutes}} Minuten, bis das Skript fertig ist.',
      timedOut:
        'Das Skript lief länger als {{minutes}} Minuten und die Seite wartet nicht mehr. Es läuft möglicherweise noch auf dem Board.'
    },
    terminal: {
      invalidBaud: 'Diese Baudrate wird nicht unterstützt.',
      invalidPort: 'Einen Gerätepfad unter /dev eingeben, zum Beispiel /dev/ttyS1.',
      invalidSettings:
        'Ungültige Einstellungen für den seriellen Port. Dies ist die Shell des Boards.',
      disconnected: 'Getrennt. Enter drücken, um neu zu verbinden.',
      title: 'Terminal',
      nanokvm: 'IronKVM Terminal',
      serial: 'Serieller Anschluss Terminal',
      serialPort: 'Serieller Anschluss',
      serialPortPlaceholder: 'Bitte seriellen Anschluss angeben',
      baudrate: 'Baudrate',
      parity: 'Parität',
      parityNone: 'Keine',
      parityEven: 'Gerade',
      parityOdd: 'Ungerade',
      flowControl: 'Fluss-Kontrolle',
      flowControlNone: 'Keine',
      flowControlSoft: 'Software',
      flowControlHard: 'Hardware',
      dataBits: 'Daten bits',
      stopBits: 'Stopp bits',
      confirm: 'Ok'
    },
    wol: {
      no: 'Nein',
      yes: 'Ja',
      deleteConfirm: 'Diese gespeicherte Adresse löschen?',
      delete: 'Löschen',
      wake: 'Aufwecken',
      rename: 'Umbenennen',
      showMac: 'MAC-Adresse anzeigen',
      showName: 'Namen anzeigen',
      requestFailed: 'Das Gerät war nicht erreichbar, der Befehl wurde nicht gesendet',
      deleteFailed: 'Löschen fehlgeschlagen',
      renameFailed: 'Umbenennen fehlgeschlagen',
      title: 'Wake-on-LAN',
      sending: 'Sende Befehl...',
      sent: 'Befehl gesendet',
      input: 'Bitte MAC Adresse eingeben',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'Hochladen fehlgeschlagen',
      uploadSuccess: 'Hochladen abgeschlossen',
      uploading: 'Wird hochgeladen: {{file}}',
      downloadingPercent: 'Wird heruntergeladen ({{percent}}): {{file}}',
      downloading: 'Wird heruntergeladen: {{file}}',
      title: 'Systemabbild Downloader',
      input: 'Bitte geben Sie die URL für das Remote-Systemabbild ein',
      ok: 'Ok',
      disabled:
        '/data Partition ist nur-lesbar, daher kann das Systemabbild nicht heruntergeladen werden',
      uploadbox: 'Datei hier ablegen oder klicken zum Auswählen',
      inputfile: 'Bitte geben Sie die Datei für das Systemabbild an',
      NoISO: 'Keine ISO',
      sha256: 'SHA-256 (optional)',
      sha256Placeholder: 'Geben Sie eine 64-stellige SHA-256-Prüfsumme ein',
      invalidSHA256: 'SHA-256 muss eine 64-stellige Hexadezimalzeichenfolge sein',
      failed: 'Download fehlgeschlagen',
      success: 'Download erfolgreich',
      checksumFailed: 'Download fehlgeschlagen: SHA-256-Prüfung fehlgeschlagen',
      cancel: 'Abbrechen',
      cancelFailed: 'Download konnte nicht abgebrochen werden',
      bootMenu: 'Bootmenü (netboot.xyz)',
      bootMenuPresent: '{{file}} liegt bereits mit korrekter Prüfsumme auf dem Gerät',
      bootMenuDesc: 'Das netboot.xyz-ISO mit geprüfter Prüfsumme für die virtuelle CD herunterladen'
    },
    alerts: {
      title: 'Braucht Aufmerksamkeit',
      temperature: {
        warning: 'Die Platine hat {{celsius}} °C. Prüfen Sie, ob Luft an sie herankommt.',
        critical:
          'Die Platine hat {{celsius}} °C und ist zu heiß. Sorgen Sie für Luft oder schalten Sie sie aus.'
      },
      storage: {
        warning:
          'Nur {{available}} von {{total}} frei auf {{path}}. Große Images passen eventuell nicht.',
        critical:
          'Nur {{available}} frei auf {{path}}. Uploads, Downloads und Add-on-Installationen schlagen fehl. Löschen Sie nicht mehr benötigte Images.'
      },
      vpn: '{{name}} soll beim Booten starten, läuft aber nicht. Der Fernzugriff darüber ist unterbrochen.',
      openVpn: 'VPN-Einstellungen öffnen',
      stream:
        'Der Videostream ist ausgefallen. Versuchen Sie im Menü Bildschirm einen anderen Videomodus oder laden Sie die Seite neu.'
    },
    power: {
      resetDesc: 'Startet den Host sofort neu. Nicht gespeicherte Arbeit geht verloren.',
      powerShortDesc:
        'Schaltet den Host ein oder bittet sein Betriebssystem, herunterzufahren (ACPI).',
      powerLongDesc: 'Schaltet den Host ohne Herunterfahren hart aus.',
      hddLed: 'HDD-LED',
      hddActive: 'Aktiv',
      hddIdle: 'Ruhig',
      title: 'Power',
      showConfirm: 'Bestätigung',
      showConfirmTip:
        'Vor einem kurzen Druck auf die Ein/Aus-Taste nachfragen. Reset und langer Druck fragen immer nach.',
      reset: 'Zurücksetzen',
      power: 'Power',
      powerShort: 'Power (Kurzer Klick)',
      powerLong: 'Power (Langer Klick)',
      resetConfirm: 'Reset-Aktion durchführen?',
      powerConfirm: 'Power-Aktion durchführen?',
      okBtn: 'Ja',
      cancelBtn: 'Nein',
      hostOs: 'Host-Betriebssystem',
      hostOsTip: 'Werden als USB-Tasten gesendet. Was sie bewirken, entscheidet der Host.',
      sleep: 'Ruhezustand',
      wake: 'Aufwecken',
      wakeKey: 'Mit Umschalttaste aufwecken',
      powerDown: 'Herunterfahren',
      sleepConfirm: 'Host in den Ruhezustand versetzen?',
      powerDownConfirm: 'Die Ausschalttaste an den Host senden?',
      wakeTip:
        'Ein schlafender Host ignoriert Aufwecken oft von dem Gerät, das ihn schlafen gelegt hat. Mit Umschalttaste aufwecken drückt eine Taste auf der Tastatur, die mehr Hosts annehmen.',
      led: 'Power-LED',
      ledOn: 'An',
      ledOff: 'Aus',
      ledUnknown: 'Unbekannt',
      ledConnected: 'Power-LED angeschlossen',
      ledConnectedTip:
        'Nur aktivieren, wenn der Power-LED-Anschluss des Hosts mit dem Board verbunden ist. Ohne ihn ist der Einschaltzustand unbekannt.',
      ledConnectedFailed: 'Die Power-LED-Einstellung konnte nicht gespeichert werden',
      powerLongConfirm:
        'Ein/Aus-Taste {{seconds}} s halten? Das schaltet den Strom ohne Herunterfahren ab.',
      done: 'Taste gedrückt',
      failed: 'Tastendruck fehlgeschlagen'
    },
    settings: {
      title: 'Einstellungen',
      nav: {
        system: 'System',
        network: 'Netzwerk',
        access: 'Zugang',
        integrations: 'Integrationen',
        boot: 'Boot und Medien',
        browser: 'Dieser Browser',
        search: 'Einstellung suchen',
        noMatch: 'Keine Einstellung gefunden',
        locked:
          'Ein Vorgang läuft. Andere Seiten und das Schließen sind erst danach wieder möglich.',
        vpnProvider: 'VPN-Anbieter'
      },
      mcp: {
        keyNote:
          'MCP verwendet einen eigenen API-Schlüssel, siehe unten. Schlüssel von der Seite API-Schlüssel funktionieren hier nicht.',
        title: 'MCP-Dienst',
        service: 'MCP-Fernsteuerung',
        serviceDesc:
          'Vertrauenswürdigen MCP-Clients erlauben, Tastatur und Maus zu steuern und Bildschirmfotos aufzunehmen',
        securityWarning:
          'Jeder mit diesem API-Schlüssel kann den entfernten Host steuern und dessen Bildschirm sehen. Verwenden Sie HTTPS und aktivieren Sie den Dienst nur in vertrauenswürdigen Netzwerken.',
        endpoint: 'Endpunkt',
        apiKey: 'API-Schlüssel',
        regenerateConfirmTitle: 'MCP-API-Schlüssel neu generieren?',
        regenerateConfirmDesc: 'Der aktuelle Schlüssel funktioniert dann sofort nicht mehr.',
        enableConfirmTitle: 'Externe MCP-Steuerung aktivieren?',
        enableConfirmDesc:
          'Durch Aktivieren von MCP wird PicoClaw gestoppt und jede aktive PicoClaw-Sitzung geschlossen.',
        failed: 'MCP-Aktion fehlgeschlagen',
        copyFailed: 'Kopieren fehlgeschlagen. Bitte manuell kopieren.',
        okBtn: 'Bestätigen',
        cancelBtn: 'Abbrechen',
        showKey: 'Schlüssel anzeigen',
        hideKey: 'Schlüssel verbergen',
        regenerateKey: 'Schlüssel neu erzeugen'
      },
      redfish: {
        example: 'Beispiel',
        title: 'Redfish',
        service: 'Redfish-Dienst',
        serviceDesc:
          'Die DMTF-Redfish-API für Stromsteuerung, virtuelle Medien und Status aus Werkzeugen wie redfishtool und Ansible. Beim Ausschalten werden alle Redfish-Sitzungen beendet.',
        endpoint: 'Service-Root',
        httpsOn: 'Das Board liefert HTTPS, was die meisten Redfish-Werkzeuge benötigen.',
        httpsOff:
          'Das Board liefert unverschlüsseltes HTTP. Die meisten Redfish-Werkzeuge benötigen HTTPS: Aktivieren Sie es unter „Einstellungen > Netzwerk“.',
        credentials:
          'Redfish akzeptiert die KVM-Konten per Basic-Authentifizierung oder Redfish-Sitzung sowie API-Schlüssel, die als X-Auth-Token gesendet werden. API-Schlüssel werden auf der Seite „API-Schlüssel“ verwaltet.',
        powerActions: 'Power-Aktionen',
        powerActionsDesc:
          'Die derzeit angebotenen Reset-Typen. On, ForceOff und GracefulShutdown benötigen den Einschaltzustand und werden daher nur angeboten, wenn „Power-LED angeschlossen“ im Power-Menü aktiviert ist.',
        sessions: 'Sitzungen',
        noSessions: 'Keine offenen Redfish-Sitzungen',
        created: 'Erstellt',
        lastUsed: 'Zuletzt verwendet',
        refresh: 'Aktualisieren',
        end: 'Beenden',
        endConfirmTitle: 'Diese Redfish-Sitzung beenden?',
        endConfirmDesc:
          'Ihr Token funktioniert sofort nicht mehr. Der Client muss sich erneut anmelden.',
        failed: 'Redfish-Vorgang fehlgeschlagen',
        copyFailed: 'Kopieren fehlgeschlagen. Bitte manuell kopieren.',
        okBtn: 'Bestätigen',
        cancelBtn: 'Abbrechen'
      },
      ipmi: {
        copyBeforeSave:
          'Kopieren Sie das Passwort jetzt. Nach dem Speichern kann es nicht mehr angezeigt werden.',
        noLogin:
          'IPMI ist an, aber kein aktives Konto hat ein IPMI-Passwort, daher kann sich niemand anmelden. Legen Sie unten eines fest.',
        title: 'IPMI',
        warning:
          'Die IPMI-Authentifizierung ist konstruktionsbedingt schwach. Wer das Board erreicht und einen Benutzernamen kennt, kann einen Hash des IPMI-Passworts dieses Benutzers abrufen und offline zu knacken versuchen. Verwenden Sie generierte Passwörter, schalten Sie IPMI nur in einem vertrauenswürdigen Netz ein und nutzen Sie lieber Redfish über HTTPS, wo ein Werkzeug es unterstützt.',
        service: 'IPMI über LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) auf UDP-Port 623 für Stromversorgung und Status des Hosts. IPMI 1.5 und Cipher Suite 0 werden abgelehnt. Ausschalten beendet alle IPMI-Sitzungen.',
        example: 'Beispiel',
        copyFailed: 'Kopieren fehlgeschlagen. Bitte manuell kopieren.',
        ledOn: 'Power status, on, off, soft, cycle und reset sind verfügbar.',
        ledOff:
          '„Power-LED angeschlossen“ ist im Power-Menü aus, daher ist der Einschaltzustand unbekannt. Nur „power reset“ funktioniert: status, on, off, soft und cycle werden abgelehnt.',
        accounts: 'Konten',
        accountsDesc:
          'IPMI meldet sich mit den KVM-Konten an, jedes mit eigenem IPMI-Passwort, getrennt vom Web-Passwort. Administratoren erhalten ADMINISTRATOR. Benutzer erhalten USER: Sie können den Einschaltzustand mit „-L USER“ lesen, aber nicht ändern.',
        passwordSet: 'IPMI-Passwort gesetzt',
        passwordNotSet: 'Kein IPMI-Passwort: keine Anmeldung über IPMI möglich',
        nameTooLong: 'Der Name ist länger als 16 Zeichen, was IPMI nicht erlaubt',
        accountDisabled: 'Das Konto ist deaktiviert',
        setPassword: 'Passwort setzen',
        changePassword: 'Passwort ändern',
        remove: 'Entfernen',
        removeConfirmTitle: 'IPMI-Passwort von {{user}} entfernen?',
        removeConfirmDesc:
          'Das Konto kann sich nicht mehr über IPMI anmelden, und seine IPMI-Sitzungen werden beendet.',
        passwordTitle: 'IPMI-Passwort für {{user}}',
        passwordDesc:
          '12 bis 20 druckbare ASCII-Zeichen, verschieden vom Web-Passwort. IPMI verlangt, dass das Board das Passwort in lesbarer Form speichert, verwenden Sie also eines, das nirgends sonst benutzt wird. Kopieren Sie es vor dem Speichern: Es wird nicht erneut angezeigt.',
        passwordPlaceholder: 'IPMI-Passwort',
        generate: 'Generieren',
        copy: 'Kopieren',
        save: 'Speichern',
        passwordLength: 'Verwenden Sie 12 bis 20 Zeichen.',
        passwordChars: 'Verwenden Sie nur druckbare ASCII-Zeichen.',
        saved: 'IPMI-Passwort gespeichert',
        failed: 'IPMI-Vorgang fehlgeschlagen',
        okBtn: 'Bestätigen',
        cancelBtn: 'Abbrechen'
      },
      ssh: {
        service: 'SSH-Server',
        serviceDesc: 'sshd jetzt und bei jedem Start ausführen',
        failed: 'SSH-Einstellungen konnten nicht geladen werden',
        rootDefault: 'root hat noch das Werkspasswort',
        rootEmpty: 'root hat kein Passwort',
        rootWarning:
          'Wer die Konsole oder SSH erreicht, kann sich als root anmelden. Legen Sie unter {{account}} > {{password}} ein Passwort fest: Für den Gerätebesitzer setzt es auch das root-Passwort.',
        connection: 'Verbindung',
        command: 'Als root anmelden',
        port: 'Port',
        viaVpn: 'Über {{name}}',
        notRunning: 'sshd läuft nicht. Schalten Sie den SSH-Server ein, um sich zu verbinden.',
        hostKeys: 'Fingerabdrücke der Host-Schlüssel',
        hostKeysDesc: 'Vergleichen Sie diese mit der Anzeige von ssh bei der ersten Verbindung.',
        noHostKeys: 'Noch keine Host-Schlüssel. sshd erzeugt sie beim ersten Start.',
        keys: 'Autorisierte Schlüssel',
        keysDesc:
          'Öffentliche Schlüssel, die sich als root anmelden dürfen. Sie liegen auf der Datenpartition und bleiben bei Updates erhalten.',
        noKeys: 'Noch keine autorisierten Schlüssel.',
        noComment: 'kein Kommentar',
        addPlaceholder:
          'Einen öffentlichen Schlüssel einfügen, etwa den Inhalt von ~/.ssh/id_ed25519.pub',
        add: 'Schlüssel hinzufügen',
        added: 'Schlüssel hinzugefügt',
        removed: 'Schlüssel entfernt',
        deleteConfirm: 'Diesen Schlüssel entfernen?',
        deleteConfirmDesc: 'Er kann sich nicht mehr anmelden. Offene Sitzungen bleiben bestehen.',
        invalidKey:
          'Das ist kein öffentlicher Schlüssel. Fügen Sie eine einzelne Zeile aus einer .pub-Datei ein.',
        keyOptions: 'Schlüssel mit Optionen wie command= oder from= werden hier nicht angenommen.',
        duplicateKey: 'Dieser Schlüssel ist bereits autorisiert.',
        lastKey:
          'Der letzte Schlüssel kann nicht entfernt werden, solange nur Schlüssel erlaubt sind.',
        keysOnly: 'Nur Schlüssel',
        keysOnlyDesc:
          'Anmeldung per Passwort und Keyboard-Interactive abschalten. Offene Sitzungen bleiben bestehen.',
        keysOnlyNeedsKey:
          'Fügen Sie zuerst einen autorisierten Schlüssel hinzu, sonst könnte sich niemand anmelden.',
        keysOnlyOn: 'Passwort-Anmeldung ausgeschaltet',
        keysOnlyOff: 'Passwort-Anmeldung eingeschaltet',
        notHonoured:
          'Der sshd dieses Images liest diese Einstellung nicht, die Passwort-Anmeldung bleibt aktiv.',
        reloadFailed:
          'Gespeichert, aber sshd konnte nicht neu geladen werden. Es gilt beim nächsten Start von sshd.',
        notApplied:
          'sshd nimmt noch Passwörter an. Schalten Sie den SSH-Server aus und wieder ein, um die Einstellung anzuwenden.',
        changePort: 'Ändern',
        portConfirm: 'SSH-Port auf {{port}} ändern?',
        portConfirmDesc:
          'Ihre bestehenden SSH-Sitzungen bleiben offen. Neue Verbindungen müssen Port {{port}} verwenden. Stellen Sie sicher, dass Ihre Firewall das zulässt.',
        portChanged: 'SSH-Port auf {{port}} geändert',
        portInvalid: 'Geben Sie einen Port von 1 bis 65535 ein.',
        portReserved: 'IronKVM selbst verwendet diesen Port. Wählen Sie einen anderen.',
        portInUse: 'Ein anderes Programm auf dem IronKVM lauscht bereits auf diesem Port.',
        portNotHonoured:
          'Der sshd dieses Images liest diese Einstellung nicht, der Port bleibt unverändert.'
      },
      vnc: {
        address: 'Adresse',
        certHint:
          'VeNCrypt X509Plain verwendet das selbstsignierte Zertifikat des Geräts, daher warnt der Client beim ersten Verbinden. Akzeptieren Sie es, oder speichern Sie das Zertifikat über die HTTPS-Adresse dieser Seite und übergeben Sie es TigerVNC mit -X509CA=<Datei>.',
        title: 'VNC',
        service: 'VNC-Server',
        serviceDesc:
          'Ein VNC-Client wie TigerVNC oder Remmina kann den Host anzeigen und steuern. Der Client muss die Tight-Kodierung unterstützen. Jeweils eine Sitzung.',
        credentials:
          'Melden Sie sich mit einem KVM-Konto an. Die Verbindung ist mit dem TLS-Zertifikat des Boards verschlüsselt (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'Der TCP-Port, auf dem der Server lauscht.',
        maxFps: 'Bildratenlimit',
        maxFpsDesc: 'Die höchste Zahl an Bildern pro Sekunde, die ein Client erhält.',
        vncAuth: 'Einfache VNC-Authentifizierung',
        vncAuthDesc:
          'Für Clients ohne VeNCrypt. Sie prüft ein eigenes VNC-Passwort statt eines Kontos.',
        vncAuthWarning:
          'Die einfache VNC-Authentifizierung verschlüsselt die Verbindung nicht. Jeder auf dem Netzwerkpfad kann den Bildschirm und die Tastatureingaben sehen. Verwenden Sie sie nur in einem vertrauenswürdigen Netzwerk.',
        password: 'VNC-Passwort',
        passwordSet: 'Ein Passwort ist gesetzt. Geben Sie ein neues ein, um es zu ändern.',
        passwordInvalid: 'Das VNC-Passwort muss 6 bis 8 Zeichen lang sein.',
        save: 'Speichern',
        saved: 'Einstellungen gespeichert',
        state: 'Status',
        listening: 'Lauscht auf Port {{port}}',
        notListening: 'Lauscht nicht',
        noSession: 'Keine offene Sitzung',
        client: 'Client',
        user: 'Benutzer',
        method: 'Authentifizierung',
        methodVencrypt: 'Konto über TLS',
        methodVnc: 'VNC-Passwort',
        since: 'Verbunden seit',
        resolution: 'Auflösung',
        framesSent: 'Gesendete Bilder',
        lastError: 'Die letzte Sitzung endete: {{error}}',
        refresh: 'Aktualisieren',
        disconnect: 'Trennen',
        disconnectConfirmTitle: 'VNC-Sitzung beenden?',
        disconnectConfirmDesc:
          'Der Client wird sofort getrennt, und jede gehaltene Taste und Maustaste wird losgelassen.',
        failed: 'VNC-Vorgang fehlgeschlagen',
        okBtn: 'Bestätigen',
        cancelBtn: 'Abbrechen'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Host-Watchdog',
        serviceDesc:
          'Wenn der Host laufen sollte und sich sein Bild während der Wartezeit nicht ändert oder kein HDMI-Signal anliegt, drückt das Board Reset oder schaltet den Host aus und wieder ein.',
        stillWarning:
          'Ein Host, dessen Bildschirm in den Ruhezustand geht oder dessen Bild bei der Arbeit stillsteht, wirkt aufgehängt. Schalten Sie den Bildschirm-Ruhezustand am Host aus oder geben Sie eine Ping-Adresse an.',
        ledHint:
          '„Power-LED angeschlossen“ ist im Power-Menü aus. Der Watchdog erkennt nicht, wann der Host aus ist, und behandelt ihn daher als immer eingeschaltet.',
        timeout: 'Wartezeit',
        timeoutDesc:
          'Wie lange der Host kein Lebenszeichen zeigen darf, bevor der Watchdog handelt.',
        action: 'Aktion',
        actionDesc:
          'Aus- und Einschalten hält die Power-Taste 5 Sekunden gedrückt und drückt sie dann erneut.',
        actionReset: 'Zurücksetzen',
        actionPower: 'Aus- und Einschalten',
        cooldown: 'Sperrzeit',
        cooldownDesc: 'Die kürzeste Zeit zwischen zwei Aktionen.',
        maxPerHour: 'Aktionen pro Stunde',
        maxPerHourDesc: 'Die höchste Zahl an Aktionen innerhalb einer Stunde.',
        pingHost: 'Ping-Adresse',
        pingHostDesc:
          'Die IP-Adresse des Hosts. Eine Antwort zählt als Lebenszeichen. Leer lassen, um nicht zu pingen.',
        pingHostInvalid: 'Geben Sie eine IPv4- oder IPv6-Adresse ein.',
        minutes: 'Min.',
        save: 'Speichern',
        saved: 'Gespeichert',
        state: 'Erkennung',
        status: {
          off: 'Aus',
          watching: 'Überwacht',
          hostOff: 'Host aus',
          captureOff: 'HDMI-Erfassung aus',
          cooldown: 'Sperrzeit',
          capped: 'Stundenlimit erreicht',
          acting: 'Handelt'
        },
        signal: 'HDMI-Signal',
        yes: 'Ja',
        no: 'Nein',
        led: 'Power-LED',
        on: 'An',
        off: 'Aus',
        ledNotConnected: 'Nicht angeschlossen',
        ping: 'Ping',
        pingNotSet: 'Nicht gesetzt',
        pingReply: 'Antwortet',
        pingNoReply: 'Keine Antwort',
        lastChange: 'Letzte Bildänderung',
        never: 'Nie',
        actsIn: 'Handelt in',
        actionsLastHour: 'Aktionen in der letzten Stunde',
        duration: '{{minutes}} Min. {{seconds}} s',
        log: 'Protokoll',
        noLog: 'Der Watchdog hat noch nicht gehandelt.',
        refresh: 'Aktualisieren',
        reasonFrozen: 'Das Bild hat sich nicht geändert',
        reasonNoSignal: 'Kein HDMI-Signal',
        stuckFor: 'kein Lebenszeichen seit {{duration}}',
        pressFailed: 'Der Tastendruck ist fehlgeschlagen: {{error}}',
        noScreenshot: 'Kein Screenshot',
        failed: 'Watchdog-Vorgang fehlgeschlagen',
        powerNeedsLed: 'Ein Aus- und Einschalten braucht „Power-LED angeschlossen“ im Power-Menü.',
        noLedConfirmTitle: 'Watchdog ohne Power-LED einschalten?',
        noLedConfirmDesc:
          'Das Board erkennt nicht, wann der Host aus ist, und behandelt ihn als immer eingeschaltet. Wenn Sie den Host herunterfahren, drückt der Watchdog nach Ablauf der Zeit Reset. Schließen Sie die Power-LED an, um das zu vermeiden.',
        noLedConfirmOk: 'Einschalten',
        cancel: 'Abbrechen'
      },
      media: {
        title: 'Virtuelle Medien',
        description:
          'Einrichtung für den Dialog Medien in der Werkzeugleiste. Images einbinden, hinzufügen und die Ventoy-Auswahl treffen geschieht im Dialog.',
        ejectFirst:
          'Der Ventoy-Datenträger steckt in einem Laufwerk. Werfen Sie ihn im Dialog Medien aus, um zu deinstallieren.'
      },
      netboot: {
        title: 'Netzwerkboot',
        isoDownload: 'Herunterladen',
        description:
          'Den Host über das Netzwerk booten: iPXE und ein Menü der Images auf dem KVM über die USB-Netzwerkverbindung, oder netboot.xyz per Proxy-DHCP im LAN.',
        addon: 'dnsmasq und Boot-Dateien',
        addonDesc:
          'Auf /data installiert: dnsmasq aus Alpine, iPXE und netboot.xyz aus ihren Releases, jeweils gegen ihre Prüfsumme geprüft.',
        install: 'Installieren',
        installing: 'Wird installiert. Das kann einige Minuten dauern.',
        uninstall: 'Deinstallieren',
        uninstallConfirm: 'Netzwerkboot ausschalten und dnsmasq sowie die Boot-Dateien entfernen?',
        needsData: 'Netzwerkboot braucht ein IronKVM-Image mit eingehängter /data-Partition.',
        usb: 'Über die USB-Netzwerkverbindung',
        usbDesc:
          'Solange die USB-Netzwerkverbindung an ist, bedient dnsmasq sie statt udhcpd. Der Host erhält seine eine Adresse ohne Router und ohne DNS-Server, iPXE für seine Architektur und ein Menü der ISO-Images auf dem KVM.',
        linkOff: 'Die USB-Netzwerkverbindung ist aus. Schalte sie unter Gerät, USB-Netzwerk ein.',
        menuUrl: 'Menü',
        leases: 'Lease des Hosts',
        noLeases: 'Noch keine',
        netbootxyzNote:
          'netboot.xyz im Menü lädt aus dem Internet, das die USB-Verbindung nicht erreicht. Der Host braucht dafür Internet an einem anderen Netzwerkanschluss.',
        lan: 'Proxy-DHCP im LAN',
        lanDesc:
          'Beantwortet PXE-Clients im LAN mit netboot.xyz, das sein Menü dann aus dem Internet lädt. Es vergibt nie Adressen und stellt die Images auf dem KVM nicht bereit.',
        lanWarning:
          'Jedem PXE-Client in diesem LAN wird netboot.xyz angeboten, nicht nur dem Host. Schalte das nur in einem Netzwerk ein, das du kontrollierst.',
        lanConfirm: 'Proxy-DHCP im LAN einschalten?',
        lanInterface: 'LAN',
        running: 'Läuft',
        stopped: 'Läuft nicht',
        images: 'Images im Menü',
        noImages: 'Keine ISO-Images im Image-Verzeichnis.',
        boots: 'Letzte Boots',
        noBoots: 'Der Host hat noch nichts abgerufen.',
        log: 'dnsmasq-Protokoll',
        refresh: 'Aktualisieren',
        okBtn: 'Bestätigen',
        cancelBtn: 'Abbrechen',
        failed: 'Netzwerkboot-Vorgang fehlgeschlagen'
      },
      about: {
        title: 'Über IronKVM',
        information: 'Informationen',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Applikations-Version',
        applicationTip: 'IronKVM Web Applikations-Version',
        image: 'Systemabbild-Version',
        imageTip: 'IronKVM-Kartenabbild und das NanoKVM-Systemabbild, auf dem es aufbaut',
        kernel: 'Kernel-Version',
        kernelTip: 'Version des derzeit laufenden Linux-Kernels',
        deviceKey: 'Geräteschlüssel',
        videoMemory: 'Videospeicher',
        videoMemoryTip:
          'Für die Videoaufnahme reservierter Speicher. Er wird nicht mit dem restlichen System geteilt.',
        videoMemoryGenerations_one: '{{count}} frühere IronKVM-Sitzung belegt Videospeicher',
        videoMemoryGenerations_other: '{{count}} frühere IronKVM-Sitzungen belegen Videospeicher',
        videoMemoryReboot: 'Neu starten, um ihn freizugeben.',
        community: 'Community',
        hostname: 'Hostname',
        hostnameUpdated: 'Hostname aktualisiert. Neustarten um zu übernehmen.',
        ipType: {
          Wired: 'Kabel',
          Wireless: 'Drahtlos',
          Other: 'Andere'
        },
        hostnameInvalid:
          'Buchstaben, Ziffern und Bindestriche verwenden, bis zu 63 pro durch Punkte getrenntem Teil. Kein Bindestrich am Anfang oder Ende eines Teils.',
        hostnameFailed: 'Der Hostname konnte nicht geändert werden',
        editHostname: 'Hostnamen bearbeiten',
        docs: 'Dokumentation',
        hardware: 'Hardware',
        hardwareFaq: 'Hardware-FAQ',
        disclaimer:
          'IronKVM: gehärtete Community-Firmware für den Sipeed NanoKVM. Nicht mit Sipeed verbunden.',
        basedOn: 'basiert auf NanoKVM {{version}}'
      },
      preferences: {
        title: 'Präferenzen'
      },
      performance: {
        title: 'Leistung'
      },
      appearance: {
        thisBrowser: 'Dieser Browser',
        thisBrowserDesc:
          'Nur in diesem Browser gespeichert. Andere Browser haben eigene Einstellungen.',
        deviceWide: 'Gerät',
        deviceWideDesc: 'Auf dem Gerät gespeichert. Gilt für alle, die es öffnen.',
        language: 'Sprache',
        languageDesc: 'Wählen Sie die Sprache für die Benutzeroberfläche aus',
        webTitle: 'Web Titel',
        webTitleDesc: 'Passen Sie den Web-Seite Titel an',
        menuBar: {
          title: 'Menüleiste',
          mode: 'Anzeigemodus',
          modeDesc: 'Menüleiste auf dem Bildschirm anzeigen',
          modeOff: 'Aus',
          modeAuto: 'Automatisch ausblenden',
          modeAlways: 'Immer sichtbar',
          keyboardLedStatus: 'Tastensperren-Anzeigen',
          keyboardLedStatusDesc:
            'Num-Lock-, Feststell- und Rollen-Status des Remote-Computers anzeigen',
          icons: 'Untermenüsymbole',
          iconsDesc: 'Untermenüsymbole in der Menüleiste anzeigen'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Tastensperren-Status der Remote-Tastatur',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num-Taste',
        numLockShort: 'Num',
        capsLock: 'Feststelltaste',
        capsLockShort: 'Fest',
        scrollLock: 'Rollen-Taste',
        scrollLockShort: 'Roll',
        on: 'Ein',
        off: 'Aus',
        unknown: 'Unbekannt'
      },
      device: {
        title: 'Gerät',
        oled: {
          title: 'OLED',
          description: 'Schalte OLED Bildschirm aus nach',
          brightness: 'OLED-Helligkeit',
          brightnessDescription: 'Eine niedrigere Stufe verlängert die Lebensdauer des Displays',
          brightnessLevels: {
            '64': 'Am niedrigsten',
            '96': 'Niedrig',
            '128': 'Mittel',
            '160': 'Hoch',
            '207': 'Standard',
            '255': 'Maximal'
          },
          0: 'Nie',
          15: '15 Sek',
          30: '30 Sek',
          60: '1 Min',
          180: '3 Min',
          300: '5 Min',
          600: '10 Min',
          1800: '30 Min',
          3600: '1 Stunde'
        },
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Frontpanel'
        },
        cpuFreq: {
          title: 'CPU-Frequenz',
          description: 'CPU-Takt für den nächsten Start festlegen',
          tip: 'Die CPU startet mit 850 MHz und ist für 1000 MHz spezifiziert. Ein neuer Wert wird beim nächsten Start übernommen, nicht im laufenden Betrieb. 1000 MHz liegt innerhalb der Spezifikation; die Temperatur bleibt bei beiden Einstellungen deutlich im zulässigen Bereich.',
          running: 'Aktuell: {{mhz}} MHz',
          rebootToApply: 'Neustart zum Übernehmen',
          rebootConfirm: 'Jetzt neu starten, um {{mhz}} MHz zu übernehmen?'
        },
        swap: {
          title: 'Swap',
          disable: 'Deaktivieren',
          description: 'Grösse der Swap-Datei festlegen',
          tip: 'Das Aktivieren dieser Funktion kann die Lebensdauer Ihrer SD-Karte verkürzen!',
          active: 'Aktiv - {{used}} von {{total}}',
          inactive: 'Eingerichtet, aber nicht in Gebrauch'
        },
        zram: {
          title: 'Komprimierter Swap (zram)',
          description: 'Swap in komprimiertem RAM statt auf der SD-Karte',
          tip: 'zram hält den Swap von der SD-Karte fern und verursacht daher keinen Verschleiß. Dahinter liegt kein Swap auf der Festplatte: Ist zram voll, beendet der Kernel einen Prozess, statt langsam auszulagern. Das Speicherlimit begrenzt, wie viel RAM zram belegen darf.',
          unavailable: 'Die Kernel-Module sind auf diesem Gerät nicht installiert',
          inactive: 'Aktiviert, aber das Gerät ist nicht gestartet',
          active: 'Aktiv - {{used}} von {{total}}, {{ratio}}x',
          off: 'Aus',
          detail: {
            algorithm: 'Algorithmus: {{algorithm}}',
            memory: 'Belegter Speicher: {{used}} von {{limit}}',
            memoryNoLimit: 'Belegter Speicher: {{used}}, kein Limit gesetzt',
            counters:
              'Seiten eingelagert {{in}}, ausgelagert {{out}} (alle Swap-Geräte, seit dem Start)'
          }
        },
        mouseJiggler: {
          title: 'Mausaktivitäts-Simulator',
          description: 'Verhindert, dass der remote Host in den Energiesparmodus wechselt',
          disable: 'Deaktivieren',
          absolute: 'Absoluter Modus',
          relative: 'Relativer Modus'
        },
        mdns: {
          description: 'mDNS-Erkennungsdienst aktivieren',
          tip: 'Deaktivieren Sie den Dienst, wenn Sie ihn nicht benötigen'
        },
        hdmi: {
          description: 'HDMI/Monitor-Ausgabe aktivieren',
          idleTimeoutTitle: 'Zeitlimit für inaktive Aufnahme',
          idleTimeoutDescription:
            'HDMI-Aufnahme stoppen, wenn keine aktiven Zuschauer vorhanden sind für',
          minutes: 'Min.'
        },
        hidOnly: 'HID-Only Mode',
        hidOnlyDesc:
          'Hören Sie auf, virtuelle Geräte zu emulieren, und behalten Sie nur die grundlegende HID-Steuerung bei',
        disk: 'Virtuelle Festplatte',
        diskDesc: 'Binde das virtuelle U-Laufwerk an den entfernten Host',
        network: 'Virtuelles Netzwerk',
        networkDesc: 'Binde die virtuelle Netzwerkkarte an den entfernten Host',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Host:',
          description:
            'Eine private Netzwerkverbindung zum entfernten Host über das USB-Kabel. Der Host erhält eine Adresse ohne Gateway und ohne DNS und erreicht Ihr LAN daher nicht über IronKVM.',
          mode: 'Protokoll',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (für Hosts ohne NCM)',
          rndis: 'RNDIS (nicht mehr angeboten)',
          rndisNote:
            'Diese Verbindung nutzt RNDIS, das nicht mehr angeboten wird. Wählen Sie NCM oder ECM.',
          subnet: 'Subnetz',
          subnetDesc:
            'Ein privates IPv4-Netz, /24 bis /30. IronKVM erhält die erste Adresse, der Host die zweite.',
          invalidSubnet: 'Geben Sie ein Subnetz wie 172.31.255.0/30 ein.',
          apply: 'Übernehmen',
          confirm: 'USB-Gerät neu verbinden?',
          reenumerate:
            'Beim Übernehmen wird die USB-Verbindung neu aufgebaut. Der Host verliert Tastatur, Maus und virtuelle Festplatte für einige Sekunden.'
        },
        audio: 'Virtueller Lautsprecher',
        audioDesc:
          'Stellt dem entfernten Host eine USB-Soundkarte bereit, damit Sie ihn hören können. Der Host muss sie als Ausgabegerät auswählen. Beim Umschalten wird die USB-Verbindung neu aufgebaut.',
        audioNote: 'Audio ist in beiden H.264-Modi (WebRTC und Direct) verfügbar, nicht in MJPEG',
        console: 'Serielle Konsole',
        consoleDesc:
          'Stellt dem entfernten Host eine serielle USB-Schnittstelle bereit, um sich bei diesem IronKVM anzumelden, wenn das Netzwerk nicht erreichbar ist',
        consoleTip:
          'Jeder, der den entfernten Host kontrolliert, erhält eine Anmeldeaufforderung dieses IronKVM. Setzen Sie vor dem Aktivieren ein starkes Passwort (Konto - Passwort ändern).',
        usbApply: {
          changed: 'Geändert',
          discard: 'Verwerfen',
          pending: 'Die Änderungen sind noch nicht übernommen.'
        },
        endpoints: {
          title: 'USB-Endpunkte',
          used: '{{used}} von {{total}} belegt',
          cost: 'belegt {{cost}}',
          needs: 'benötigt {{cost}}',
          full: 'Nicht genügend USB-Endpunkte. Schalten Sie zuerst etwas anderes aus.',
          inactive:
            'Aktiviert, aber nicht aktiv: Dem USB-Controller sind die Endpunkte ausgegangen. Schalten Sie ein anderes Gerät aus, dann startet dieses sofort.',
          explain:
            'Der USB-Controller hat eine feste Anzahl eingehender Endpunkte, und diese werden hier gezählt. Sind mehr Geräte aktiviert als hineinpassen, bleiben Tastatur und Maus erhalten und der Rest wird ausgeschaltet.',
          error: 'Das Gerät ist nicht erreichbar. Bitte erneut versuchen.',
          fitTogether: 'Diese passen zusammen: {{sets}}'
        },
        reboot: 'Neustarten',
        rebootDesc: 'Sind Sie sicher dass Sie IronKVM neustarten möchten?',
        okBtn: 'Ja',
        cancelBtn: 'Nein',
        rebootFailed: 'Neustart fehlgeschlagen'
      },
      network: {
        title: 'Netzwerk',
        wifi: {
          disconnectBtn: 'Trennen',
          disconnectWarning:
            'Wenn Sie IronKVM über dieses WLAN erreichen, verliert diese Seite ihre Verbindung.',
          disconnected: 'WLAN getrennt',
          title: 'Wi-Fi',
          description: 'Wi-Fi konfigurieren',
          apMode: 'AP-Modus ist aktiviert, verbinden Sie sich per QR-Code mit dem Wi-Fi',
          connect: 'Wi-Fi verbinden',
          connectDesc1: 'Bitte geben Sie die Netzwerk-SSID und das Passwort ein',
          connectDesc2: 'Bitte geben Sie das Passwort ein, um diesem Netzwerk beizutreten',
          disconnect: 'Möchten Sie die Netzwerkverbindung wirklich trennen?',
          failed: 'Verbindung fehlgeschlagen, bitte erneut versuchen.',
          ssid: 'Name',
          password: 'Passwort',
          joinBtn: 'Verbinden',
          confirmBtn: 'OK',
          cancelBtn: 'Abbrechen'
        },
        tls: {
          description: 'HTTPS-Protokoll aktivieren',
          tip: 'Hinweis: Die Verwendung von HTTPS kann die Latenz erhöhen, besonders im MJPEG-Videomodus.',
          restarting: 'Der Geräteserver wird neu gestartet, das dauert etwa zwei Minuten...',
          waiting: 'Warten, bis das Gerät wieder antwortet...',
          waitingHttp:
            'Wechsel zurück zu http. Laden Sie diese Seite neu, falls sie sich nicht von selbst öffnet.',
          failed: 'Die HTTPS-Einstellung konnte nicht geändert werden',
          enableConfirm: 'HTTPS einschalten?',
          disableConfirm: 'HTTPS ausschalten?',
          confirmDesc:
            'Sie werden abgemeldet und der Geräteserver startet neu, was etwa zwei Minuten dauert. Danach öffnet die Seite {{url}}.',
          confirmOk: 'Fortfahren',
          confirmCancel: 'Abbrechen'
        },
        ethernet: {
          title: 'IP-Adresse',
          description:
            'Legen Sie fest, wie IronKVM seine Adresse im kabelgebundenen Netzwerk erhält',
          dhcp: 'DHCP',
          manual: 'Manuell',
          networkDetails: 'Netzwerkdetails',
          interface: 'Schnittstelle',
          ipAddress: 'IP-Adresse',
          subnetMask: 'Subnetzmaske',
          router: 'Router',
          save: 'Übernehmen',
          invalidAddress: 'Bitte geben Sie eine gültige IP-Adresse ein',
          invalidMask:
            'Bitte geben Sie eine gültige Subnetzmaske ein, zum Beispiel 255.255.255.0 oder 24',
          invalidRouter: 'Bitte geben Sie eine gültige Router-Adresse ein',
          addressRequired: 'Eine IP-Adresse ist erforderlich',
          maskRequired: 'Eine Subnetzmaske ist erforderlich',
          applyTitle: 'Adresse von IronKVM ändern?',
          applyWarning:
            'Die Verbindung zu dieser Seite geht verloren. IronKVM übernimmt die neue Adresse und wartet {{seconds}} Sekunden darauf, dass Sie es dort erreichen. Das Erreichen behält die Änderung. Erreicht es nichts, stellt IronKVM die vorherigen Einstellungen wieder her.',
          applyConfirm: 'Übernehmen',
          applyCancel: 'Abbrechen',
          applyFailed: 'Die Adresse konnte nicht übernommen werden',
          trialTitle: 'Warten auf Bestätigung',
          trialDhcp: 'IronKVM fordert eine Adresse per DHCP an.',
          trialStatic: 'IronKVM ist jetzt unter {{address}} erreichbar.',
          trialInstruction:
            'Öffnen Sie IronKVM unter seiner neuen Adresse und melden Sie sich an, falls es danach fragt. Damit bleibt die Änderung erhalten. Erreicht IronKVM innerhalb von {{seconds}} Sekunden nichts, stellt es die vorherigen Einstellungen wieder her.',
          trialOpen: 'Neue Adresse öffnen',
          trialKeep: 'Diese Einstellungen behalten',
          trialKept: 'Die neue Adresse ist gespeichert',
          trialKeepFailed: 'Die Einstellungen konnten nicht behalten werden',
          trialGone: 'Die Änderung wurde bereits zurückgenommen. Bitte erneut versuchen.',
          unsaved: 'Ungespeicherte Änderungen'
        },
        dns: {
          title: 'DNS',
          description: 'DNS-Server für IronKVM konfigurieren',
          mode: 'Modus',
          dhcp: 'DHCP',
          manual: 'Manuell',
          add: 'DNS hinzufügen',
          save: 'Speichern',
          invalid: 'Bitte geben Sie eine gültige IP-Adresse ein',
          noDhcp: 'Derzeit ist kein DHCP-DNS verfügbar',
          saved: 'DNS-Einstellungen gespeichert',
          saveFailed: 'DNS-Einstellungen konnten nicht gespeichert werden',
          unsaved: 'Ungespeicherte Änderungen',
          maxServers: 'Maximal {{count}} DNS-Server erlaubt',
          dnsServers: 'DNS-Server',
          dhcpServersDescription: 'DNS-Server werden automatisch per DHCP bezogen',
          manualServersDescription: 'DNS-Server können manuell bearbeitet werden',
          networkDetails: 'Netzwerkdetails',
          interface: 'Schnittstelle',
          ipAddress: 'IP-Adresse',
          subnetMask: 'Subnetzmaske',
          router: 'Router',
          none: 'Keine'
        }
      },
      vpn: {
        connect: 'Verbinden',
        connectDesc:
          'Mit dem {{name}}-Netzwerk verbinden. Aus trennt die Verbindung, ohne den Dienst zu beenden.',
        kvmUrl: 'KVM-Adresse',
        moreTip: 'Weitere Aktionen',
        restartTip: 'Neu starten',
        stopTip: 'Stoppen',
        updateTip: 'Auf {{version}} aktualisieren',
        loading: 'Lädt...',
        okBtn: 'Ja',
        cancelBtn: 'Nein',
        restart: '{{name}} neu starten?',
        stop: '{{name}} stoppen?',
        stopDesc:
          'Der Dienst wird jetzt gestoppt. „Beim Systemstart starten“ ist ein eigener Schalter und bleibt unverändert.',
        update: '{{name}} auf {{version}} aktualisieren?',
        updateDesc: 'Der Dienst wird neu gestartet, falls er läuft. Die Anmeldung bleibt erhalten.',
        notInstall: '{{name}} ist nicht installiert.',
        install: 'Installieren',
        installing: 'Wird installiert',
        installFailed: 'Installation fehlgeschlagen',
        retry: 'Erneut versuchen',
        notRunning: '{{name}} läuft nicht. Starten Sie es, um fortzufahren.',
        run: 'Starten',
        boot: 'Beim Systemstart starten',
        bootDesc: '{{name}} beim Hochfahren des KVM starten.',
        control: 'Steuerungsserver',
        connected: 'Verbunden',
        disconnected: 'Nicht verbunden',
        deviceName: 'Gerätename',
        deviceIP: 'Geräte-IP',
        account: 'Konto',
        version: 'Version',
        uptime: 'Laufzeit',
        peers: 'Peers',
        noPeers: 'Noch keine Peers.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Speicher',
        daemonRss: 'Dienst',
        group: 'Add-on-Gruppe',
        high: 'gedrosselt über {{size}}',
        max: 'vom Kernel beendet über {{size}}',
        noGroup: 'Keine Speichergruppe für Add-ons auf diesem Board.',
        uninstall: '{{name}} deinstallieren',
        uninstallDesc:
          'Möchten Sie {{name}} wirklich deinstallieren? Die Anmeldung bleibt auf dem Board gespeichert.',
        blocked:
          '{{other}} läuft oder startet beim Systemstart. Es kann nur ein VPN gleichzeitig laufen: Stoppen Sie zuerst {{other}} und deaktivieren Sie dessen Start beim Systemstart.',
        swap: {
          title: 'Swap-Speicher',
          tip: 'Wenn dem Dienst der Speicher knapp wird, versuchen Sie, Swap zu aktivieren. Das geht unter „Einstellungen > Leistung“.'
        },
        copy: 'Kopieren',
        copied: 'Link kopiert',
        copyFailed: 'Link konnte nicht kopiert werden. Markieren und von Hand kopieren.',
        open: 'Öffnen',
        checkAgain: 'Erneut prüfen',
        notSignedIn:
          'Noch nicht angemeldet. Die Anmeldung über den Link abschließen und erneut prüfen.',
        checkFailed: 'Anmeldestatus konnte nicht geprüft werden',
        loginWaiting:
          'Diese Seite prüft alle paar Sekunden und macht weiter, sobald Sie angemeldet sind.',
        uninstallFailed: 'Deinstallation fehlgeschlagen',
        loginFailed: 'Anmeldung fehlgeschlagen'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Laden Sie das',
        package: 'Installations-Paket herunter',
        unzip: 'und entpacken Sie es',
        notLogin:
          'Das Gerät konnte noch nicht gefunden werden. Bitte melden Sie sich an und verknüpfen Sie dieses Gerät mit Ihrem Konto.',
        urlPeriod: 'Diese URL ist für 10 Minuten gültig',
        login: 'Anmelden',
        logout: 'Abmelden',
        logoutDesc: 'Möchten Sie sich wirklich abmelden?',
        manualIntro: 'Oder per SSH von Hand installieren:',
        copyBinaries: 'tailscale und tailscaled nach {{dir}} auf dem IronKVM kopieren',
        linksFile: 'Im selben Verzeichnis eine Datei namens links mit diesen zwei Zeilen anlegen:',
        rebootRefresh: 'Den IronKVM neu starten, dann diese Seite neu laden'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Dieses Gerät ist noch keinem NetBird-Netzwerk beigetreten. Treten Sie mit einem Setup-Key bei oder melden Sie sich per SSO an.',
        setupKey: 'Setup-Key',
        setupKeyPlaceholder: 'Setup-Key aus dem NetBird-Dashboard einfügen',
        join: 'Beitreten',
        or: 'oder',
        sso: 'Mit SSO anmelden',
        urlPeriod: 'Diese URL ist für 10 Minuten gültig',
        logout: 'Abmelden',
        logoutDesc:
          'Beim Abmelden wird dieser Peer aus Ihrem NetBird-Konto entfernt und seine Konfiguration hier gelöscht. Ein erneuter Beitritt erfordert einen Setup-Key oder eine SSO-Anmeldung, und der Peer erhält möglicherweise eine neue IP. Fortfahren?',
        joinFailed: 'Beitritt zum Netzwerk fehlgeschlagen'
      },
      update: {
        title: 'Nach Aktualisierungen suchen',
        queryFailed: 'Version konnte nicht abgefragt werden',
        updateFailed: 'Aktualisierung fehlgeschlagen. Bitte versuchen Sie es erneut.',
        isLatest: 'Sie haben bereits die aktuellste Version.',
        available: 'Eine Aktualisierung ist verfügbar. Möchten sie diese jetzt durchführen?',
        updating: 'Aktualisierung gestartet. Bitte warten...',
        confirm: 'Bestätigen',
        cancel: 'Abbrechen',
        preview: 'Vorab-Versionen',
        previewDesc: 'Erhalten Sie vorab Zugriff auf neue Funktionen und Verbesserungen',
        previewTip:
          'Bitte beachten Sie, dass Vorab-Versionen womöglich noch Fehler oder unvollständige Funktionen enthalten!',
        customServer: {
          title: 'Benutzerdefinierter Update-Server',
          desc: 'Online-Updates von einem angegebenen Server suchen und herunterladen',
          invalidUrl:
            'Geben Sie ein gültiges HTTP- oder HTTPS-Serververzeichnis ohne Abfrageparameter, Fragment oder latest.json ein.',
          loadFailed: 'Die Konfiguration des Update-Servers konnte nicht geladen werden.',
          saveFailed: 'Die Konfiguration des Update-Servers konnte nicht gespeichert werden.',
          saved: 'Die Konfiguration des Update-Servers wurde gespeichert.',
          save: 'Speichern',
          confirmTitle: 'Benutzerdefinierten Update-Server verwenden?',
          confirmDesc:
            'SHA-512 bestätigt lediglich, dass das Paket mit dem von diesem Server bereitgestellten Manifest übereinstimmt. Es beweist nicht, dass das Paket eine offizielle IronKVM-Version ist. Ein fehlerhafter oder bösartiger Server kann das Gerät unbrauchbar machen, Datenverlust verursachen oder das System kompromittieren.',
          confirm: 'Trotzdem verwenden',
          useSipeed: 'Offiziellen Sipeed-Server verwenden',
          previewDisabled:
            'Vorschau-Updates sind nicht verfügbar, solange ein benutzerdefinierter Update-Server aktiviert ist.'
        },
        offline: {
          chooseFile: 'Datei wählen',
          installing: 'Hochgeladen. Wird installiert...',
          noFile: 'Keine Datei gewählt',
          title: 'Offline Aktualisierung',
          desc: 'Über lokales Installationspaket aktualisieren',
          upload: 'Hochladen',
          checksumPlaceholder: 'SHA-256-Prüfsumme (optional)',
          invalidChecksum: 'Die SHA-256-Prüfsumme muss 64 hexadezimale Zeichen enthalten.',
          checksumMismatch:
            'Die SHA-256-Überprüfung ist fehlgeschlagen. Das Paket ist möglicherweise beschädigt.',
          invalidName:
            'Ungültiges Dateinamenformat. Bitte laden Sie von den GitHub-Releases herunter.',
          updateFailed: 'Aktualisierung fehlgeschlagen. Bitte versuchen Sie es erneut.'
        },
        updateTo: 'Auf {{version}} aktualisieren',
        updateConfirmDesc:
          'Das Gerät installiert das Update und startet seinen Server neu. Diese Seite lädt neu, sobald der Server wieder da ist.',
        releaseNotes: 'Versionshinweise'
      },
      account: {
        title: 'Konto',
        webAccount: 'Web Konto Name',
        role: 'Rolle',
        roles: { admin: 'Administrator', user: 'Benutzer' },
        password: 'Passwort',
        updateBtn: 'Ändern',
        logoutBtn: 'Abmelden',
        logoutDesc: 'Möchten Sie sich wirklich abmelden?',
        okBtn: 'Ja',
        cancelBtn: 'Nein',
        users: {
          title: 'Benutzer',
          create: 'Benutzer anlegen',
          enabled: 'Aktiviert',
          disabled: 'Deaktiviert',
          deviceOwner: 'Gerätebesitzer',
          resetPassword: 'Passwort zurücksetzen',
          delete: 'Löschen',
          deleteConfirm: 'Diesen Benutzer löschen und alle seine Sitzungen widerrufen?',
          created: 'Benutzer angelegt',
          deleted: 'Benutzer gelöscht',
          passwordUpdated: 'Passwort aktualisiert',
          loadFailed: 'Benutzer konnten nicht geladen werden',
          saveFailed: 'Benutzer konnte nicht gespeichert werden',
          deleteFailed: 'Benutzer konnte nicht gelöscht werden'
        }
      },
      apiKeys: {
        mcpNote:
          'Diese Schlüssel funktionieren nicht für MCP. MCP hat einen eigenen Schlüssel auf der MCP-Seite.',
        metricsUrl: 'Metrik-URL',
        monitoring: 'Überwachung',
        monitoringDesc:
          'Prometheus liest die Metriken mit einem API-Schlüssel von dieser Seite, gesendet als Bearer-Token. Jede Rolle darf sie lesen.',
        scrapeConfig: 'Prometheus-Scrape-Konfiguration',
        title: 'API-Schlüssel',
        description:
          'Ein Schlüssel handelt im Namen seines Besitzers, mit dessen Rolle. Senden Sie ihn als Authorization: Bearer <key> für Metriken und die API oder als X-Auth-Token für Redfish.',
        name: 'Name',
        namePlaceholder: 'Wofür der Schlüssel ist, z. B. prometheus',
        nameRequired: 'Geben Sie dem Schlüssel einen Namen',
        nameTooLong: 'Der Name darf höchstens 64 Zeichen lang sein',
        unnamed: '(unbenannt)',
        create: 'Schlüssel erstellen',
        created: 'Erstellt',
        owner: 'Besitzer',
        empty: 'Keine API-Schlüssel',
        newKeyTitle: 'Ihr neuer API-Schlüssel',
        newKeyWarning:
          'Kopieren Sie den Schlüssel jetzt. Er wird nicht gespeichert und kann nicht erneut angezeigt werden. Wenn Sie ihn verlieren, widerrufen Sie ihn und erstellen Sie einen neuen.',
        copy: 'Kopieren',
        copied: 'Kopiert',
        copyFailed: 'Kopieren fehlgeschlagen. Bitte manuell kopieren.',
        done: 'Fertig',
        revoke: 'Widerrufen',
        revokeConfirmTitle: 'Diesen API-Schlüssel widerrufen?',
        revokeConfirmDesc: 'Alles, was „{{name}}“ verwendet, funktioniert sofort nicht mehr.',
        revoked: 'API-Schlüssel widerrufen',
        loadFailed: 'API-Schlüssel konnten nicht geladen werden',
        createFailed: 'API-Schlüssel konnte nicht erstellt werden',
        revokeFailed: 'API-Schlüssel konnte nicht widerrufen werden',
        cancelBtn: 'Abbrechen'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistent',
      empty: 'Öffnen Sie das Bedienfeld und starten Sie eine Aufgabe.',
      inputPlaceholder: 'Beschreiben Sie, was der PicoClaw tun soll',
      newConversation: 'Neues Gespräch',
      processing: 'Wird verarbeitet...',
      agent: {
        defaultTitle: 'Allgemeiner Assistent',
        defaultDescription: 'Allgemeine Chat-, Such- und Arbeitsbereichshilfe.',
        kvmTitle: 'Fernsteuerung',
        kvmDescription: 'Betreiben Sie den Remote-Host über IronKVM.',
        switched: 'Agentenrolle gewechselt',
        switchFailed: 'Agentenrolle konnte nicht gewechselt werden'
      },
      send: 'Senden',
      cancel: 'Abbrechen',
      status: {
        connecting: 'Verbindung zum Gateway wird hergestellt...',
        connected: 'PicoClaw Sitzung verbunden',
        disconnected: 'PicoClaw Sitzung geschlossen',
        stopped: 'Stoppanforderung gesendet',
        runtimeStarted: 'PicoClaw Runtime gestartet',
        runtimeStartFailed: 'PicoClaw Runtime konnte nicht gestartet werden',
        runtimeStopped: 'PicoClaw Runtime gestoppt',
        runtimeStopFailed: 'PicoClaw Runtime konnte nicht gestoppt werden',
        controlSwitchedToMCP: 'Steuerung zum externen MCP-Dienst gewechselt'
      },
      connection: {
        runtime: {
          checking: 'Überprüfung',
          restoring: 'PicoClaw wird wiederhergestellt',
          ready: 'Runtime bereit',
          stopped: 'Runtime gestoppt',
          blockedByMCP: 'Externe MCP-Steuerung ist aktiv',
          readyBlockedByMCP:
            'Die Runtime läuft, aber externes MCP steuert gerade die Geräteeingabe.',
          readyWithoutControl:
            'Die Runtime läuft. Erteile PicoClaw die Gerätesteuerung, bevor du neu verbindest.',
          unavailable: 'Runtime nicht verfügbar',
          configError: 'Konfigurationsfehler'
        },
        transport: {
          connecting: 'Verbinden',
          connected: 'Verbunden',
          disconnected: 'Getrennt',
          reconnect: 'Neu verbinden',
          reconnectDescription: 'Mit der laufenden PicoClaw-Sitzung neu verbinden.',
          reconnectBlocked: 'PicoClaw braucht die Gerätesteuerung, bevor es neu verbinden kann.'
        },
        run: {
          idle: 'Leerlauf',
          busy: 'Beschäftigt'
        }
      },
      message: {
        toolAction: 'Aktion',
        observation: 'Beobachtung',
        screenshot: 'Screenshot'
      },
      overlay: {
        locked: 'PicoClaw steuert das Gerät. Die manuelle Eingabe wird angehalten.'
      },
      control: {
        picoclaw: 'Gerätesteuerung: PicoClaw',
        picoclawDescription:
          'PicoClaw kann Tastatur- und Mauseingaben senden. Manuelle Eingaben können pausieren.',
        mcp: 'Gerätesteuerung: externes MCP',
        mcpDescription:
          'Externes MCP kann auf das Gerät schreiben. PicoClaw übernimmt die Eingabe nicht.',
        off: 'Gerätesteuerung: aus',
        offDescription:
          'Die KI sendet keine Tastatur- oder Mauseingaben. Die manuelle Steuerung bleibt verfügbar.',
        transitioning: 'Gerätesteuerung: wird gewechselt',
        transitioningDescription: 'Die Gerätesteuerung wird abgeglichen. Bitte warten.',
        grant: 'Steuerung erteilen',
        release: 'Freigeben',
        releasing: 'Wird freigegeben...',
        switching: 'Wird gewechselt...',
        releasingLabel: 'Gerätesteuerung: wird freigegeben',
        releasingDescription:
          'Die Gerätesteuerung wird zurückgegeben. PicoClaw hat laufende Eingaben beendet.',
        granted: 'PicoClaw-Steuerung erteilt',
        released: 'PicoClaw-Steuerung freigegeben',
        grantFailed: 'PicoClaw-Steuerung konnte nicht erteilt werden',
        releaseFailed: 'PicoClaw-Steuerung konnte nicht freigegeben werden',
        grantConfirmTitle: 'Gerätesteuerung zu PicoClaw wechseln?',
        grantConfirmDesc: 'Externe MCP-Geräteschreibvorgänge werden unterbrochen.'
      },
      install: {
        install: 'Installieren Sie PicoClaw',
        installing: 'Installation von PicoClaw',
        success: 'PicoClaw erfolgreich installiert',
        failed: 'Installation von PicoClaw fehlgeschlagen',
        uninstalling: 'Runtime wird deinstalliert...',
        uninstalled: 'Runtime erfolgreich deinstalliert.',
        uninstallFailed: 'Deinstallation fehlgeschlagen.',
        requiredTitle: 'PicoClaw ist nicht installiert',
        requiredDescription: 'Installieren Sie PicoClaw, bevor Sie die PicoClaw Runtime starten.',
        progressDescription: 'PicoClaw wird heruntergeladen und installiert.',
        stages: {
          preparing: 'Vorbereiten',
          downloading: 'Wird heruntergeladen',
          extracting: 'Extrahieren',
          verifying: 'Überprüfen',
          installing: 'Installiere',
          installed: 'Installiert',
          install_timeout: 'Zeitüberschreitung',
          install_failed: 'Fehlgeschlagen'
        }
      },
      model: {
        requiredTitle: 'Modellkonfiguration ist erforderlich',
        requiredDescription:
          'Konfigurieren Sie das PicoClaw-Modell, bevor Sie den PicoClaw-Chat verwenden.',
        docsTitle: 'Konfigurationshandbuch',
        docsDesc: 'Unterstützte Modelle und Protokolle',
        menuLabel: 'Modell konfigurieren',
        modelIdentifier: 'Modell-ID',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API-Schlüssel',
        apiKeyPlaceholder: 'API-Schlüssel des Modells eingeben',
        save: 'Speichern',
        saving: 'Speichern',
        saved: 'Modellkonfiguration gespeichert',
        saveFailed: 'Modellkonfiguration konnte nicht gespeichert werden',
        invalid: 'Modellkennung, API Base URL und API-Schlüssel sind erforderlich'
      },
      uninstall: {
        menuLabel: 'Deinstallieren',
        confirmTitle: 'Deinstallieren PicoClaw',
        confirmContent:
          'Sind Sie sicher, dass Sie PicoClaw deinstallieren möchten? Dadurch werden die ausführbare Datei und alle Konfigurationsdateien gelöscht.',
        confirmOk: 'Deinstallieren',
        confirmCancel: 'Abbrechen'
      },
      history: {
        title: 'Verlauf',
        loading: 'Sitzungen werden geladen...',
        emptyTitle: 'Noch keine Historie',
        emptyDescription: 'Frühere PicoClaw-Sitzungen werden hier angezeigt.',
        loadFailed: 'Der Sitzungsverlauf konnte nicht geladen werden',
        deleteFailed: 'Sitzung konnte nicht gelöscht werden',
        deleteConfirmTitle: 'Sitzung löschen',
        deleteConfirmContent: 'Sind Sie sicher, dass Sie „{{title}}“ löschen möchten?',
        deleteConfirmOk: 'Löschen',
        deleteConfirmCancel: 'Abbrechen',
        messageCount_one: '{{count}} Nachricht',
        messageCount_other: '{{count}} Nachrichten',
        messageCount: '{{count}} Nachrichten'
      },
      config: {
        startRuntime: 'PicoClaw starten',
        stopRuntime: 'PicoClaw stoppen'
      },
      start: {
        enableConfirmTitle: 'Steuerung zu PicoClaw wechseln?',
        enableConfirmDesc: 'Beim Starten von PicoClaw wird der externe MCP-Dienst deaktiviert.',
        enableConfirmOk: 'PicoClaw starten',
        enableConfirmCancel: 'Abbrechen',
        title: 'PicoClaw starten',
        description:
          'Starten Sie die Runtime, um mit der Verwendung des PicoClaw-Assistenten zu beginnen.',
        switchFromMCP: 'Zu PicoClaw wechseln und starten',
        takeoverAndStart: 'Übernehmen und starten'
      }
    },
    error: {
      title: 'Wir sind auf ein Problem gestossen',
      refresh: 'Neuladen',
      panel: 'Dieser Teil der Seite funktioniert nicht mehr',
      retry: 'Erneut versuchen'
    },
    fullscreen: {
      toggle: 'Vollbild ein/aus'
    },
    input: {
      disconnected: 'Tastatur und Maus sind nicht verbunden',
      disconnectedTls:
        'Der Browser hat die sichere Verbindung für Tastatur und Maus abgelehnt, und zwar ohne nachzufragen. Das von diesem Gerät erzeugte Zertifikat wird noch nicht als vertrauenswürdig eingestuft. Öffnen Sie diese Adresse in einem neuen Tab, akzeptieren Sie das Zertifikat und laden Sie dann neu. Zuverlässig hilft nur die Installation des Zertifikats.',
      disconnectedNever:
        'Die Verbindung für Tastatur und Maus konnte nicht aufgebaut werden. Der Rest der Seite funktioniert, weil er sie nicht nutzt. Prüfen Sie, ob etwas zwischen Ihnen und dem Gerät sie blockiert.',
      disconnectedDropped:
        'Die Verbindung für Tastatur und Maus ist abgebrochen und nicht wiederhergestellt. Nach einem Neustart verbindet sie sich von selbst; bleibt diese Meldung bestehen, laden Sie die Seite neu.',
      hidDisabled: 'HID ist auf diesem Gerät ausgeschaltet (/boot/disable_hid).',
      keyFailed: 'Die Taste konnte nicht gesendet werden.'
    },
    speaker: { title: 'Lautsprecher', unmute: 'Ton an', mute: 'Stummschalten' },
    upstream: {
      check: 'Nach Updates suchen',
      updateTo: 'Auf {{version}} aktualisieren',
      confirm: '{{name}} auf {{version}} aktualisieren?',
      confirmDesc:
        'Die neue Version wird von GitHub geladen und mit den dort veröffentlichten Prüfsummen geprüft. Schlägt etwas fehl, bleibt die aktuelle Version erhalten.',
      ok: 'Aktualisieren',
      upToDate: 'Aktuell',
      builtIn: 'mitgeliefert',
      checkFailed: 'Suche nach Updates fehlgeschlagen: {{error}}',
      unverifiable: 'Version {{version}} wird nicht angeboten: {{reason}}',
      inUse: 'Aktualisierung gerade nicht möglich: {{reason}}',
      running: 'Aktualisiere auf {{version}}...',
      done: '{{name}} auf {{version}} aktualisiert',
      failed: 'Die letzte Aktualisierung ist fehlgeschlagen: {{error}}'
    },
    menu: {
      mediaAdd: 'Image hinzufügen',
      mediaMoreOptions: 'Weitere Optionen',
      mediaSettings: 'Medien-Einstellungen',
      collapse: 'Menü einklappen',
      expand: 'Menü ausklappen',
      more: 'Mehr',
      media: 'Medien',
      tools: 'Werkzeuge',
      text: 'Text',
      advanced: 'Erweitert',
      mediaMounted: 'Eingebunden',
      mediaLibrary: 'Bibliothek',
      textToHost: 'Zum Host',
      textFromHost: 'Vom Host'
    },
    ion: {
      checking: 'Videospeicher wird vor dem Start des Streams geprüft...',
      warn: 'Der Videospeicher wird knapp. Ein einziger Neustart des Servers würde ihn erschöpfen. Starten Sie bei Gelegenheit neu.',
      criticalTitle: 'Nicht genug Videospeicher, um den Stream zu starten',
      criticalBody:
        'Das Starten des Videos würde den reservierten Speicher erschöpfen und den Server stoppen. Alle anderen Funktionen arbeiten weiterhin, einschließlich Stromsteuerung und Neustart. Nur ein Neustart des IronKVM gibt diesen Speicher wieder frei.',
      criticalContinue: 'Video trotzdem starten',
      criticalReboot: 'IronKVM neu starten',
      criticalRebooting: 'Wird neu gestartet...'
    }
  }
};

export default de;
