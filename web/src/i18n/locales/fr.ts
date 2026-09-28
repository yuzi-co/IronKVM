const fr = {
  translation: {
    head: {
      desktop: 'Bureau à distance',
      login: 'Connexion',
      changePassword: 'Changer le mot de passe',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        "Le navigateur a refusé d'enregistrer la session. Un cookie laissé par une session HTTPS précédente ne peut pas être remplacé en http simple. Effacez les cookies de cette adresse, ou ouvrez une fenêtre privée, puis reconnectez-vous.",
      login: 'Connexion',
      placeholderUsername: "Veuillez entrer votre nom d'utilisateur",
      placeholderPassword: 'Veuillez entrer votre mot de passe',
      placeholderCurrentPassword: 'Mot de passe actuel',
      placeholderPassword2: 'Veuillez entrer votre mot de passe à nouveau',
      noEmptyUsername: "Le nom d'utilisateur ne peut pas être vide",
      noEmptyPassword: 'Le mot de passe ne peut pas être vide',
      passwordLength: 'Le mot de passe doit contenir entre 8 et 72 caractères',
      noAccount:
        "Impossible de récupérer les informations de l'utilisateur, veuillez rafraîchir la page ou réinitialiser le mot de passe",
      invalidUser: "Nom d'utilisateur ou mot de passe invalide",
      locked: 'Trop de connexions, veuillez réessayer plus tard',
      globalLocked: 'Système sous protection, veuillez réessayer plus tard',
      error: 'Erreur inattendue',
      invalidCurrentPassword: 'Le mot de passe actuel est incorrect',
      changePassword: 'Changer le mot de passe',
      changePasswordDesc:
        'Pour la sécurité de votre appareil, veuillez modifier le mot de passe de connexion Web.',
      differentPassword: 'Les mots de passe ne correspondent pas',
      illegalUsername: "Le nom d'utilisateur contient des caractères illégaux",
      illegalPassword: 'Le mot de passe contient des caractères illégaux',
      forgetPassword: 'Mot de passe oublié',
      ok: 'Se connecter',
      cancel: 'Annuler',
      loginButtonText: 'Connexion',
      tips: {
        reset1:
          'Pour réinitialiser les mots de passe, appuyez et maintenez enfoncé le bouton BOOT sur le NanoKVM pendant 10 secondes.',
        reset2: 'Pour les étapes détaillées, veuillez consulter ce document :',
        reset3: 'Compte Web par défaut :',
        reset4: 'Compte SSH par défaut :',
        change1: 'Veuillez noter que cette action modifiera les mots de passe suivants :',
        change2: 'Mot de passe de connexion Web',
        change3: 'Mot de passe racine du système (mot de passe de connexion SSH)',
        change4:
          'Pour réinitialiser les mots de passe, appuyez et maintenez enfoncé le bouton BOOT sur le NanoKVM.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Configurez le Wi-Fi pour le NanoKVM',
      success:
        'Veuillez vérifier le statut du réseau du NanoKVM et visitez la nouvelle adresse IP.',
      failed: "L'opération a échoué, veuillez réessayer.",
      invalidMode:
        'Le mode actuel ne prend pas en charge la configuration réseau. Veuillez accéder à votre appareil et activer le mode de configuration Wi-Fi.',
      confirmBtn: 'Ok',
      finishBtn: 'Terminé',
      ap: {
        authTitle: 'Authentification requise',
        authDescription: 'Veuillez saisir le mot de passe AP pour continuer',
        authFailed: 'Mot de passe AP invalide',
        passPlaceholder: 'AP mot de passe',
        verifyBtn: 'Vérifier'
      }
    },
    screen: {
      scale: 'Échelle',
      title: 'Écran',
      video: 'Mode vidéo',
      videoDirectTips: 'Activez HTTPS dans "Paramètres > Appareil" pour utiliser ce mode',
      resolution: 'Résolution',
      controlRegion: {
        title: 'Étalonnage de la souris',
        description:
          "Utilisez ce réglage lorsque l'appareil contrôlé utilise une résolution autre que 16:9 et que le curseur est décalé horizontalement ou verticalement.",
        off: 'Désactivé',
        auto: 'Automatique',
        autoWarning:
          "L'étalonnage peut échouer si l'application utilisateur présente un arrière-plan entièrement noir.",
        manual: 'Manuel',
        selectedResolution: 'Résolution de la zone sélectionnée',
        unused: 'Non utilisée',
        originalResolution: "Résolution d'origine",
        selectResolution: "Sélectionner la résolution d'origine",
        addResolution: 'Ajouter une résolution personnalisée',
        add: 'Ajouter',
        duplicateResolution: 'Cette résolution existe déjà.',
        width: 'Largeur',
        height: 'Hauteur',
        apply: 'Calculer et appliquer',
        invalidResolution: "Saisissez une résolution d'origine valide une fois la vidéo prête.",
        select: 'Sélectionner une zone',
        clear: 'Rétablir la détection automatique',
        saveFailed: "Échec de l'enregistrement de la zone d'entrée.",
        tooSmall: 'La zone sélectionnée est trop petite.',
        previewUnavailable: 'Aperçu indisponible',
        clearConfirm: 'Rétablir la détection automatique des bordures noires ?',
        dragHint: 'Faites glisser pour sélectionner la zone du bureau distant',
        finish: 'Terminé',
        confirm: 'Confirmer',
        cancel: 'Annuler'
      },
      auto: 'Automatique',
      autoTips:
        "Sous certaines résolutions, il peut y avoir des artefacts visuels ou un décalage de la souris. Veuillez ajuster la résolution de l'hôte distant ou désactiver le mode automatique.",
      fps: 'FPS',
      customizeFps: 'Personnaliser',
      quality: 'Qualité',
      qualityLossless: 'Sans perte',
      qualityHigh: 'Élevé',
      qualityMedium: 'Moyen',
      qualityLow: 'Bas',
      frameDetect: 'Détection de trame',
      frameDetectTip:
        "Calcule la différence entre les images. Arrête la transmission du flux vidéo lorsqu'aucun changement n'est détecté sur l'écran de l'hôte distant",
      resetHdmi: 'Réinitialiser le HDMI',
      mixedH264: {
        title: 'Conflit de flux H.264',
        description:
          'Les modes H.264 Direct et H.264 WebRTC sont utilisés simultanément. Cela peut provoquer des déchirures d’écran ou une vidéo corrompue. Veuillez n’utiliser qu’un seul mode H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Échec de la connexion WebRTC',
        description: 'Vérifiez la connexion réseau ou changez de mode vidéo.'
      },
      captureStatus: {
        hdmiError: 'Erreur d’image HDMI',
        unsupportedResolution: 'La résolution actuelle n’est pas prise en charge',
        retrieving: 'Récupération de l’image...',
        changingResolution: 'Changement de résolution...',
        updateFailed: 'L’image ne peut pas être mise à jour pour le moment',
        videoError: 'Erreur d’affichage vidéo',
        noHdmi: 'Aucun signal HDMI détecté',
        unavailable: 'L’image ne peut pas être affichée pour le moment'
      }
    },
    keyboard: {
      title: 'Clavier',
      paste: 'Coller',
      tips: "Tape le texte sur l'hôte sous forme de frappes de touches. Choisissez la disposition de clavier de l'hôte.",
      placeholder: 'Veuillez saisir',
      submit: 'Soumettre',
      virtual: 'Clavier',
      readClipboard: 'Lire le presse-papiers',
      clipboardPermissionDenied:
        "Accès au presse-papiers refusé. Veuillez autoriser l'accès dans votre navigateur.",
      clipboardReadError: 'Échec de la lecture du presse-papiers',
      mediaKeys: {
        title: 'Touches multimédia',
        mute: 'Muet',
        volumeDown: 'Baisser le volume',
        volumeUp: 'Monter le volume',
        previous: 'Piste précédente',
        playPause: 'Lecture ou pause',
        next: 'Piste suivante',
        stop: 'Arrêt'
      },
      pasting: {
        layout: "Disposition du clavier de l'hôte",
        layouts: {
          us: 'Anglais (États-Unis)',
          uk: 'Anglais (Royaume-Uni)',
          de: 'Allemand',
          fr: 'Français',
          es: 'Espagnol',
          it: 'Italien',
          ptBr: 'Portugais (Brésil)',
          se: 'Suédois / finnois',
          ru: 'Russe',
          ja: 'Japonais',
          ko: 'Coréen'
        },
        speed: 'Vitesse de frappe',
        speeds: {
          fast: 'Rapide',
          normal: 'Normale',
          slow: 'Lente'
        },
        estimate: 'Durée de frappe : environ {{duration}}',
        untypeable: 'Caractères que cette disposition ne peut pas taper : {{count}}',
        untypeableAt: 'ligne {{line}}, colonne {{column}}',
        skipUntypeable: 'Taper le reste',
        shortcut: "{{shortcut}} tape directement le presse-papiers sur l'hôte.",
        clipboardUnavailable:
          "Le navigateur ne laisse une page lire le presse-papiers qu'en HTTPS. Collez le texte dans la zone avec Ctrl+V.",
        clipboardEmpty: 'Le presse-papiers ne contient pas de texte.',
        tooLong: 'Le texte est trop long. La limite est de {{max}} caractères.',
        inProgress: 'Un collage est déjà en cours de frappe.',
        typing: "Frappe sur l'hôte",
        done: 'Texte tapé',
        canceled: 'Collage annulé',
        failed: 'Échec du collage',
        cancel: 'Annuler',
        controlBusy: 'Un autre contrôleur utilise le clavier.',
        hidError: "Les frappes n'ont pas pu être envoyées à l'hôte."
      },
      shortcut: {
        title: 'Raccourcis',
        custom: 'Personnalisé',
        capture: 'Cliquez ici pour capturer le raccourci',
        clear: 'Effacer',
        save: 'Enregistrer',
        captureTips:
          'La capture de touches système (comme la touche Windows) nécessite l’autorisation du plein écran.',
        enterFullScreen: 'Basculer en mode plein écran.'
      },
      leaderKey: {
        title: 'Touche Leader',
        desc: "Contournez les restrictions du navigateur et envoyez les raccourcis système directement à l'hôte distant.",
        howToUse: 'Comment utiliser',
        simultaneous: {
          title: 'Mode simultané',
          desc1: 'Maintenez la touche Leader enfoncée, puis appuyez sur le raccourci.',
          desc2: 'Intuitif, mais peut entrer en conflit avec les raccourcis système.'
        },
        sequential: {
          title: 'Mode séquentiel',
          desc1:
            'Appuyez sur la touche Leader → appuyez sur le raccourci dans l’ordre → appuyez à nouveau sur la touche Leader.',
          desc2: "Nécessite plus d'étapes, mais évite complètement les conflits système."
        },
        enable: 'Activer la touche Leader',
        tip: 'Lorsqu’elle est définie comme touche Leader, cette touche sert uniquement de déclencheur de raccourci et perd son comportement par défaut.',
        placeholder: 'Appuyez sur la touche Leader',
        shiftRight: 'Shift droit',
        ctrlRight: 'Ctrl droit',
        metaRight: 'Win droit',
        submit: 'Soumettre',
        recorder: {
          rec: 'REC',
          activate: 'Activer les touches',
          input: 'Veuillez appuyer sur le raccourci...'
        }
      }
    },
    mouse: {
      title: 'Souris',
      cursor: 'Style de curseur',
      default: 'Curseur par défaut',
      pointer: 'Curseur de la souris',
      cell: 'Curseur de cellule',
      text: 'Curseur de texte',
      grab: 'Curseur de poignée',
      hide: 'Cacher le curseur',
      mode: 'Mode de la souris',
      absolute: 'Mode absolu',
      relative: 'Mode relatif',
      absoluteShort: 'Absolu',
      relativeShort: 'Relatif',
      absoluteStalled: 'La cible ignore la souris absolue',
      absoluteStalledDesc:
        "La cible ne lit plus les rapports de souris absolue, les déplacements du pointeur sont donc perdus. Le clavier n'est pas affecté. Réinitialiser l'USB règle souvent le problème ; le mode relatif utilise un autre endpoint.",
      useRelative: 'Passer en mode relatif',
      direction: 'Sens de la molette',
      scrollUp: 'Faire défiler vers le haut',
      scrollDown: 'Faites défiler vers le bas',
      speed: 'Vitesse de la molette',
      fast: 'Rapide',
      slow: 'Lent',
      requestPointer:
        'Pour utiliser le mode relatif, cliquez sur le bureau pour capturer le pointeur de la souris.',
      resetHid: 'Réinitialiser le périphérique HID',
      hidOnly: {
        title: 'Mode HID uniquement',
        desc: "Si votre souris et votre clavier ne répondent plus et que la réinitialisation de HID ne vous aide pas, il peut s'agir d'un problème de compatibilité entre le NanoKVM et l'appareil. Essayez d'activer le mode HID-Only pour une meilleure compatibilité.",
        tip1: "L'activation du mode HID-Only démontera le disque U virtuel et le réseau virtuel",
        tip2: "En mode HID-Only, le montage d'image est désactivé",
        rebuild: 'Changer de mode reconstruit la connexion USB. NanoKVM ne redémarre pas',
        enable: 'Activer le mode HID uniquement',
        disable: 'Désactiver le mode HID uniquement'
      }
    },
    image: {
      title: 'Images',
      loading: 'Chargement',
      empty: 'Vide',
      mountMode: 'Mode de montage',
      mountFailed: "Échec du montage de l'image.",
      mountDesc:
        "Dans certains systèmes, il est nécessaire de déséjecter le disque virtuel sur l'hôte distant avant de monter l'image.",
      unmountFailed: 'Échec du démontage',
      unmountDesc:
        "Sur certains systèmes, vous devez l'éjecter manuellement de l'hôte distant avant de démonter l'image.",
      refresh: 'Actualiser la liste des images',
      disk: 'Disque',
      cdrom: 'CD',
      driveEmpty: 'Vide',
      eject: 'Éjecter',
      readOnly: 'Lecture seule',
      readOnlyTip: "S'applique à la prochaine image insérée dans le disque.",
      noDrives: 'Aucun lecteur virtuel. Activez le disque virtuel dans les Paramètres.',
      insertFailed: "Échec de l'insertion",
      ejectFailed: "Échec de l'éjection",
      insertInto: 'Sera insérée dans : {{drive}}. Cliquez pour changer.',
      loadedIn: 'Dans le lecteur : {{drive}}',
      attention: 'Attention',
      deleteConfirm: 'Etes-vous sûr de vouloir supprimer cette image?',
      okBtn: 'Oui',
      cancelBtn: 'Non',
      deleteFailed: 'Échec de la suppression',
      ventoy: {
        statusNoKernel: 'Non pris en charge par ce firmware',
        statusNotInstalled: 'Non installé',
        statusReady: 'Prêt',
        statusSelected: 'Images sélectionnées : {{count}}',
        statusInDrive: 'Dans le lecteur de disque, {{size}}',
        noKernel:
          "Le noyau de ce firmware ne prend pas en charge device-mapper, Ventoy ne peut donc pas être utilisé tant qu'une image qui le prend en charge n'est pas installée.",
        installDesc: "Démarrez l'hôte depuis plusieurs images sur un seul disque, sans les copier.",
        install: 'Installer',
        installing: 'Téléchargement de Ventoy, environ 20 Mo. Cela peut prendre quelques minutes.',
        needsData: 'Ventoy nécessite une image IronKVM avec la partition /data montée.',
        uninstall: 'Désinstaller',
        uninstallConfirm: 'Supprimer les fichiers de Ventoy ?',
        noImages: 'Aucune image à placer sur le disque Ventoy.',
        onDisk: 'Sur le disque Ventoy',
        missing: 'Manquant : {{file}}',
        remove: 'Retirer du disque Ventoy',
        setHint:
          "La sélection d'images ne change que lorsque le disque Ventoy n'est dans aucun lecteur.",
        useAsDisk: 'Utiliser comme disque virtuel',
        failed: 'Échec de la requête Ventoy',
        secureBoot:
          "Avec Secure Boot activé, l'hôte doit enregistrer une fois la clé de Ventoy dans MokManager. Le fichier de clé ENROLL_THIS_KEY_IN_MOKMANAGER.cer se trouve sur la partition VTOYEFI.",
        readOnly:
          "L'hôte voit le disque en lecture seule, donc la persistance Ventoy et ventoy.json sur le lecteur ne fonctionnent pas."
      },
      tips: {
        title: 'Comment télécharger',
        usb1: 'Connectez le NanoKVM à votre ordinateur via USB.',
        usb2: 'Assurez-vous que le disque virtuel est monté (Paramètres - Disque virtuel).',
        usb3: 'Ouvrez le disque virtuel sur votre ordinateur et copiez le fichier image dans le répertoire racine du disque virtuel.',
        scp1: 'Assurez-vous que le NanoKVM et votre ordinateur sont sur le même réseau local.',
        scp2: 'Ouvrez un terminal sur votre ordinateur et utilisez la commande SCP pour copier le fichier image dans le répertoire /data du NanoKVM.',
        scp3: 'Exemple : scp chemin-de-votre-image root@ip-de-votre-nanokvm:/data',
        tfCard: 'Carte TF',
        tf1: 'Cette méthode est adaptée aux systèmes Linux.',
        tf2: 'Retirez la carte TF du NanoKVM (Pour la version FULL, il est nécessaire de retirer le boîtier).',
        tf3: 'Insérez la carte TF dans un lecteur de carte et connectez-la à votre ordinateur.',
        tf4: 'Copiez le fichier image dans le répertoire /data de la carte TF sur votre ordinateur.',
        tf5: 'Réinsérez la carte TF dans le NanoKVM.'
      }
    },
    script: {
      title: 'Script',
      upload: 'Téléverser',
      run: 'Exécuter',
      runBackground: 'Exécuter en arrière-plan',
      runFailed: "Échec de l'exécution",
      attention: 'Attention',
      delDesc: 'Êtes-vous sûr de vouloir supprimer ce fichier ?',
      confirm: 'Oui',
      cancel: 'Non',
      delete: 'Supprimer',
      close: 'Fermer'
    },
    terminal: {
      title: 'Terminal',
      nanokvm: 'Terminal NanoKVM',
      serial: 'Terminal Port Série',
      serialPort: 'Port série',
      serialPortPlaceholder: 'Veuillez entrer le port série',
      baudrate: 'Débit en bauds',
      parity: 'Parité',
      parityNone: 'Aucun',
      parityEven: 'Paire',
      parityOdd: 'Impaire',
      flowControl: 'Contrôle de débit',
      flowControlNone: 'Aucun',
      flowControlSoft: 'Logiciel',
      flowControlHard: 'Matériel',
      dataBits: 'Bits de données',
      stopBits: "Bits d'arrêt",
      confirm: 'Ok'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'Envoi de la commande...',
      sent: 'Commande envoyée',
      input: "Veuillez entrer l'adresse MAC",
      ok: 'Ok'
    },
    download: {
      title: 'Télécharger l’image',
      input: 'Veuillez entrer l’URL d’une image distante',
      ok: 'Ok',
      disabled: 'La partition /data est en lecture seule, impossible de télécharger l’image',
      uploadbox: 'Déposez le fichier ici ou cliquez pour sélectionner',
      inputfile: 'Veuillez saisir le fichier image',
      NoISO: 'Aucun ISO',
      sha256: 'SHA-256 (facultatif)',
      sha256Placeholder: 'Saisissez une somme de contrôle SHA-256 de 64 caractères',
      invalidSHA256: 'SHA-256 doit être une chaîne hexadécimale de 64 caractères',
      failed: 'Échec du téléchargement',
      success: 'Téléchargement réussi',
      checksumFailed: 'Échec du téléchargement : échec de la vérification SHA-256',
      cancel: 'Annuler',
      cancelFailed: 'Impossible d’annuler le téléchargement',
      bootMenu: 'Menu de démarrage (netboot.xyz)',
      bootMenuDesc: "Télécharger l'ISO netboot.xyz, somme de contrôle vérifiée, pour le CD virtuel"
    },
    power: {
      title: 'Power',
      showConfirm: 'Confirmation',
      showConfirmTip: 'Les opérations électriques nécessitent une confirmation supplémentaire',
      reset: 'Réinitialiser',
      power: 'Power',
      powerShort: 'Power (appui court)',
      powerLong: 'Power (appui long)',
      resetConfirm: "Procéder à l'opération de réinitialisation?",
      powerConfirm: 'Continuer le fonctionnement électrique?',
      okBtn: 'Oui',
      cancelBtn: 'Non',
      hostOs: "OS de l'hôte",
      hostOsTip: "Envoyées comme touches USB. L'hôte décide de leur effet.",
      sleep: 'Veille',
      wake: 'Réveil',
      wakeKey: 'Réveil avec Maj',
      powerDown: 'Éteindre',
      sleepConfirm: "Mettre l'hôte en veille ?",
      powerDownConfirm: "Envoyer la touche d'extinction à l'hôte ?",
      wakeTip:
        "Un hôte en veille ignore souvent Réveil venant de l'appareil qui l'a mis en veille. Réveil avec Maj appuie sur une touche du clavier, que davantage d'hôtes acceptent.",
      led: "LED d'alimentation",
      ledOn: 'Allumée',
      ledOff: 'Éteinte',
      ledUnknown: 'Inconnu',
      ledConnected: "LED d'alimentation branchée",
      ledConnectedTip:
        "N'activez cette option que si le connecteur de LED d'alimentation de l'hôte est câblé à la carte. Sans cela, l'état d'alimentation est inconnu.",
      ledConnectedFailed: "Impossible d'enregistrer le réglage de la LED d'alimentation"
    },
    settings: {
      title: 'Paramètres',
      mcp: {
        title: 'Service MCP',
        service: 'Contrôle à distance MCP',
        serviceDesc:
          'Autoriser les clients MCP de confiance à contrôler le clavier et la souris et à prendre des captures d’écran',
        securityWarning:
          'Toute personne possédant cette clé API peut contrôler l’hôte distant et voir son écran. Utilisez HTTPS et activez ce service uniquement sur des réseaux de confiance.',
        endpoint: 'Point de terminaison',
        apiKey: 'Clé API',
        regenerateConfirmTitle: 'Régénérer la clé API MCP ?',
        regenerateConfirmDesc: 'La clé actuelle cessera immédiatement de fonctionner.',
        enableConfirmTitle: 'Activer le contrôle MCP externe ?',
        enableConfirmDesc:
          'L’activation de MCP arrêtera PicoClaw et fermera toute session PicoClaw active.',
        failed: 'Échec de l’opération MCP',
        copyFailed: 'La copie a échoué. Copiez manuellement.',
        okBtn: 'Confirmer',
        cancelBtn: 'Annuler'
      },
      redfish: {
        title: 'Redfish',
        service: 'Service Redfish',
        serviceDesc:
          "L'API Redfish de la DMTF, pour le contrôle de l'alimentation, les médias virtuels et l'état, depuis des outils comme redfishtool et Ansible. La désactiver met fin à toutes les sessions Redfish.",
        endpoint: 'Racine du service',
        httpsOn: 'La carte sert en HTTPS, ce dont la plupart des outils Redfish ont besoin.',
        httpsOff:
          'La carte sert en HTTP simple. La plupart des outils Redfish ont besoin de HTTPS : activez-le dans "Paramètres > Réseau".',
        credentials:
          "Redfish accepte les comptes du KVM, avec l'authentification Basic ou une session Redfish, ainsi que les clés API envoyées en X-Auth-Token. Les clés API se gèrent sur la page Clés API.",
        powerActions: "Actions d'alimentation",
        powerActionsDesc:
          "Les types de réinitialisation proposés actuellement. On, ForceOff et GracefulShutdown nécessitent l'état d'alimentation : ils ne sont proposés que si \"LED d'alimentation branchée\" est activé dans le menu Power.",
        sessions: 'Sessions',
        noSessions: 'Aucune session Redfish ouverte',
        created: 'Créée',
        lastUsed: 'Dernière utilisation',
        refresh: 'Actualiser',
        end: 'Terminer',
        endConfirmTitle: 'Terminer cette session Redfish ?',
        endConfirmDesc:
          'Son jeton cesse immédiatement de fonctionner. Le client devra se reconnecter.',
        failed: "Échec de l'opération Redfish",
        copyFailed: 'Échec de la copie. Copiez manuellement.',
        okBtn: 'Confirmer',
        cancelBtn: 'Annuler'
      },
      watchdog: {
        title: 'Watchdog',
        service: "Watchdog de l'hôte",
        serviceDesc:
          "Si l'hôte devrait être allumé et que son image ne change pas, ou qu'il n'y a pas de signal HDMI, pendant le délai, la carte appuie sur reset ou éteint puis rallume l'hôte.",
        stillWarning:
          "Un hôte dont l'écran se met en veille, ou dont l'image reste fixe pendant qu'il travaille, semble bloqué. Désactivez la mise en veille de l'écran sur l'hôte, ou indiquez une adresse de ping.",
        ledHint:
          "« LED d'alimentation branchée » est désactivé dans le menu d'alimentation. Le watchdog ne voit pas quand l'hôte est éteint, il le considère donc toujours allumé.",
        timeout: 'Délai',
        timeoutDesc:
          "Durée pendant laquelle l'hôte peut ne montrer aucun signe de vie avant que le watchdog agisse.",
        action: 'Action',
        actionDesc:
          "Le cycle d'alimentation maintient le bouton d'alimentation 5 secondes, puis appuie à nouveau.",
        actionReset: 'Reset',
        actionPower: "Cycle d'alimentation",
        cooldown: 'Temps de repos',
        cooldownDesc: 'Le temps minimal entre deux actions.',
        maxPerHour: 'Actions par heure',
        maxPerHourDesc: "Le nombre maximal d'actions sur une heure.",
        pingHost: 'Adresse de ping',
        pingHostDesc:
          "L'adresse IP de l'hôte. Une réponse compte comme un signe de vie. Laissez vide pour ne pas pinguer.",
        pingHostInvalid: 'Saisissez une adresse IPv4 ou IPv6.',
        minutes: 'min',
        save: 'Enregistrer',
        saved: 'Enregistré',
        state: 'Détecteur',
        status: {
          off: 'Désactivé',
          watching: 'Surveillance',
          hostOff: 'Hôte éteint',
          captureOff: 'Capture HDMI désactivée',
          cooldown: 'Temps de repos',
          capped: 'Limite horaire atteinte',
          acting: 'En action'
        },
        signal: 'Signal HDMI',
        yes: 'Oui',
        no: 'Non',
        led: "LED d'alimentation",
        on: 'Allumée',
        off: 'Éteinte',
        ledNotConnected: 'Non connectée',
        ping: 'Ping',
        pingNotSet: 'Non défini',
        pingReply: 'Répond',
        pingNoReply: 'Pas de réponse',
        lastChange: "Dernier changement d'image",
        never: 'Jamais',
        actsIn: 'Agit dans',
        actionsLastHour: 'Actions dans la dernière heure',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Journal',
        noLog: "Le watchdog n'a encore rien fait.",
        refresh: 'Actualiser',
        reasonFrozen: "L'image n'a pas changé",
        reasonNoSignal: 'Pas de signal HDMI',
        stuckFor: 'aucun signe de vie depuis {{duration}}',
        pressFailed: "L'appui a échoué : {{error}}",
        noScreenshot: "Pas de capture d'écran",
        failed: 'Échec de l’opération du watchdog'
      },
      netboot: {
        title: 'Démarrage réseau',
        description:
          "Démarrer l'hôte depuis le réseau : iPXE et un menu des images du KVM par la liaison réseau USB, ou netboot.xyz par proxy DHCP sur le LAN.",
        addon: 'dnsmasq et fichiers de démarrage',
        addonDesc:
          'Installés sur /data : dnsmasq depuis Alpine, iPXE et netboot.xyz depuis leurs versions publiées, chacun vérifié par sa somme de contrôle.',
        install: 'Installer',
        installing: 'Installation en cours. Cela peut prendre quelques minutes.',
        uninstall: 'Désinstaller',
        uninstallConfirm:
          'Désactiver le démarrage réseau et supprimer dnsmasq et les fichiers de démarrage ?',
        needsData:
          'Le démarrage réseau nécessite une image IronKVM avec la partition /data montée.',
        usb: 'Sur la liaison réseau USB',
        usbDesc:
          "Tant que la liaison réseau USB est active, dnsmasq la sert à la place d'udhcpd. L'hôte reçoit son unique adresse sans routeur ni serveur DNS, iPXE pour son architecture et un menu des images ISO du KVM.",
        linkOff: 'La liaison réseau USB est désactivée. Activez-la dans Appareil, Réseau USB.',
        menuUrl: 'Menu',
        leases: "Bail de l'hôte",
        noLeases: 'Aucun pour le moment',
        netbootxyzNote:
          "netboot.xyz dans le menu se charge depuis Internet, que la liaison USB n'atteint pas. L'hôte a besoin d'Internet sur un autre port réseau.",
        lan: 'Proxy DHCP sur le LAN',
        lanDesc:
          "Répond aux clients PXE du LAN avec netboot.xyz, qui charge ensuite son menu depuis Internet. Il ne distribue jamais d'adresses et ne sert pas les images du KVM.",
        lanWarning:
          "netboot.xyz est proposé à tous les clients PXE de ce LAN, pas seulement à l'hôte. N'activez ceci que sur un réseau que vous contrôlez.",
        lanConfirm: 'Activer le proxy DHCP sur le LAN ?',
        lanInterface: 'LAN',
        running: 'En cours',
        stopped: 'Arrêté',
        images: 'Images dans le menu',
        noImages: 'Aucune image ISO dans le répertoire des images.',
        boots: 'Démarrages récents',
        noBoots: "L'hôte n'a encore rien récupéré.",
        log: 'Journal de dnsmasq',
        refresh: 'Actualiser',
        okBtn: 'Confirmer',
        cancelBtn: 'Annuler',
        failed: 'Échec de l’opération de démarrage réseau'
      },
      about: {
        title: 'A propos de NanoKVM',
        information: 'Informations',
        ip: 'IP',
        mdns: 'mDNS',
        application: "Version de l'application",
        applicationTip: "Version de l'application Web NanoKVM",
        image: "Version de l'image",
        imageTip: "Version de l'image système NanoKVM",
        kernel: 'Version du noyau',
        kernelTip: "Version du noyau Linux en cours d'exécution",
        deviceKey: "Clé de l'appareil",
        videoMemory: 'Mémoire vidéo',
        videoMemoryTip:
          "Mémoire réservée à la capture vidéo. Elle n'est pas partagée avec le reste du système.",
        videoMemoryGenerations_one:
          '{{count}} session NanoKVM précédente occupe de la mémoire vidéo',
        videoMemoryGenerations_other:
          '{{count}} sessions NanoKVM précédentes occupent de la mémoire vidéo',
        videoMemoryReboot: 'Redémarrez pour la récupérer.',
        community: 'Communauté',
        hostname: "Nom d'hôte",
        hostnameUpdated: "Nom d'hôte mis à jour. Redémarrez pour appliquer.",
        ipType: {
          Wired: 'Filaire',
          Wireless: 'Sans fil',
          Other: 'Autre'
        }
      },
      appearance: {
        title: 'Apparence',
        display: 'Affichage',
        language: 'Langue',
        languageDesc: "Sélectionnez la langue de l'interface",
        webTitle: 'Titre Web',
        webTitleDesc: 'Personnaliser le titre de la page Web',
        menuBar: {
          title: 'Barre de menus',
          mode: "Mode d'affichage",
          modeDesc: "Afficher la barre de menu sur l'écran",
          modeOff: 'Désactivé',
          modeAuto: 'Masquer automatiquement',
          modeAlways: 'Toujours visible',
          keyboardLedStatus: 'Indicateurs de verrouillage du clavier',
          keyboardLedStatusDesc:
            'Afficher l’état de Verr Num, Verr Maj et Arrêt défil du poste distant',
          icons: 'Icônes du sous-menu',
          iconsDesc: 'Afficher les icônes des sous-menus dans la barre de menus'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'État des verrouillages du clavier distant',
        indicatorLabel: '{{label}} : {{state}}',
        numLock: 'Verr Num',
        numLockShort: 'Num',
        capsLock: 'Verr Maj',
        capsLockShort: 'Maj',
        scrollLock: 'Arrêt défil',
        scrollLockShort: 'Défil',
        on: 'Activé',
        off: 'Désactivé',
        unknown: 'Inconnu'
      },
      device: {
        title: 'Appareil',
        oled: {
          title: 'OLED',
          description: "Écran OLED s'éteint automatiquement",
          brightness: "Luminosité de l'OLED",
          brightnessDescription: "Un niveau plus bas prolonge la durée de vie de l'écran",
          brightnessLevels: {
            '64': 'Minimale',
            '96': 'Faible',
            '128': 'Moyenne',
            '160': 'Élevée',
            '207': 'Par défaut',
            '255': 'Maximale'
          },
          0: 'Jamais',
          15: '15 sec',
          30: '30 sec',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 heure'
        },
        ssh: {
          description: "Activer l'accès à distance SSH",
          tip: "Définissez un mot de passe fort avant d'activer (Compte - Modifier le mot de passe)"
        },
        advanced: 'Paramètres avancés',
        cpuFreq: {
          title: 'Fréquence du CPU',
          description: 'Définir la fréquence du CPU appliquée au prochain démarrage',
          tip: "Le CPU démarre à 850 MHz et est spécifié pour 1000 MHz. Une nouvelle valeur s'applique au prochain démarrage, pas pendant que le système tourne. 1000 MHz reste dans les spécifications ; la température reste largement dans les limites avec les deux réglages.",
          running: 'Actuelle : {{mhz}} MHz',
          rebootToApply: 'redémarrer pour appliquer',
          rebootConfirm: 'Redémarrer maintenant pour appliquer {{mhz}} MHz ?'
        },
        swap: {
          title: 'Échange',
          disable: 'Désactiver',
          description: "Définir la taille du fichier d'échange",
          tip: "L'activation de cette fonctionnalité pourrait réduire la durée de vie de votre carte SD!"
        },
        zram: {
          title: 'Échange compressé (zram)',
          description: 'Échange dans la RAM compressée, au lieu de la carte SD',
          tip: "zram garde l'échange hors de la carte SD, il ne l'use donc pas. Il n'y a pas d'échange sur disque derrière : si zram est plein, le noyau arrête un processus au lieu de paginer lentement. La limite de mémoire plafonne la quantité de RAM que zram peut prendre.",
          unavailable: 'Les modules du noyau ne sont pas installés sur cet appareil',
          inactive: "Activé, mais le périphérique n'a pas démarré",
          active: 'Actif - {{used}} sur {{total}}, {{ratio}}x',
          off: 'Désactivé',
          detail: {
            algorithm: 'Algorithme : {{algorithm}}',
            memory: 'Mémoire utilisée : {{used}} sur {{limit}}',
            memoryNoLimit: 'Mémoire utilisée : {{used}}, aucune limite définie',
            counters:
              "Pages échangées entrantes {{in}}, sortantes {{out}} (tous les périphériques d'échange, depuis le démarrage)"
          }
        },
        mouseJiggler: {
          title: 'Souris Jiggler',
          description: "Empêcher l'hôte distant de dormir",
          disable: 'Désactiver',
          absolute: 'Mode absolu',
          relative: 'Mode relatif'
        },
        mdns: {
          description: 'Activer le service de découverte mDNS',
          tip: "L'éteindre si ce n'est pas nécessaire"
        },
        hdmi: {
          description: 'Activer HDMI/sortie moniteur',
          idleTimeoutTitle: "Délai d'inactivité de la capture",
          idleTimeoutDescription:
            "Arrêter la capture HDMI lorsqu'il n'y a aucun spectateur actif pendant",
          minutes: 'min'
        },
        autostart: {
          title: 'Paramètres des scripts de démarrage automatique',
          description: "Gérer les scripts qui s'exécutent automatiquement au démarrage du système",
          new: 'Nouveau',
          deleteConfirm: 'Êtes-vous sûr de vouloir supprimer ce fichier ?',
          yes: 'Oui',
          no: 'Non',
          scriptName: 'Nom du script de démarrage automatique',
          scriptContent: 'Contenu du script de démarrage automatique',
          settings: 'Paramètres'
        },
        hidOnly: 'HID-Mode uniquement',
        hidOnlyDesc:
          "Arrêtez d'émuler des périphériques virtuels, en ne conservant que le contrôle de base HID",
        disk: 'Disque virtuel',
        diskDesc: "Monter le disque virtuel U sur l'hôte distant",
        network: 'Réseau virtuel',
        networkDesc: "Monter la carte réseau virtuelle sur l'hôte distant",
        usbNetwork: {
          description:
            "Une liaison réseau privée avec l'hôte distant par le câble USB. L'hôte reçoit une adresse sans passerelle ni DNS, il ne peut donc pas atteindre votre réseau local via NanoKVM.",
          off: 'Désactivé',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (pour les hôtes sans NCM)',
          rndis: "RNDIS (n'est plus proposé)",
          rndisNote: "Cette liaison utilise RNDIS, qui n'est plus proposé. Choisissez NCM ou ECM.",
          subnet: 'Sous-réseau',
          subnetDesc:
            "Un réseau IPv4 privé, de /24 à /30. NanoKVM prend la première adresse, l'hôte la deuxième.",
          addresses: 'NanoKVM : {{board}}, hôte : {{host}}',
          invalidSubnet: 'Saisissez un sous-réseau tel que 172.31.255.0/30.',
          apply: 'Appliquer',
          confirm: 'Reconnecter le périphérique USB ?',
          reenumerate:
            "L'application reconstruit la connexion USB. L'hôte perd le clavier, la souris et le disque virtuel pendant quelques secondes."
        },
        audio: 'Haut-parleur virtuel',
        audioDesc:
          "Présenter une carte son USB à l'hôte distant, pour pouvoir l'entendre. L'hôte doit la choisir comme périphérique de sortie. Modifier ce réglage reconstruit la connexion USB.",
        audioNote:
          "L'audio est disponible dans les deux modes H.264 (WebRTC et Direct), pas en MJPEG",
        console: 'Console série',
        consoleDesc:
          "Présenter un port série USB à l'hôte distant, pour se connecter à ce NanoKVM quand le réseau est inaccessible",
        consoleTip:
          "Quiconque contrôle l'hôte distant obtient une invite de connexion sur ce NanoKVM. Définissez un mot de passe fort avant d'activer (Compte - Modifier le mot de passe).",
        endpoints: {
          title: 'Endpoints USB',
          used: '{{used}} sur {{total}} utilisés',
          cost: 'utilise {{cost}}',
          needs: 'nécessite {{cost}}',
          full: "Pas assez d'endpoints USB. Désactivez d'abord autre chose.",
          inactive:
            "Activé, mais inactif : le contrôleur USB n'a plus d'endpoints. Désactivez un autre périphérique et celui-ci démarre aussitôt.",
          explain:
            "Le contrôleur USB dispose d'un nombre fixe d'endpoints entrants, et c'est eux que l'on compte ici. Si plus de périphériques sont activés qu'il n'y a de place, le clavier et la souris sont conservés et le reste est désactivé.",
          error: "Impossible de joindre l'appareil. Réessayez.",
          fitTogether: 'Compatibles ensemble : {{sets}}'
        },
        reboot: 'Redémarrer',
        rebootDesc: 'Êtes-vous sûr de vouloir redémarrer NanoKVM?',
        okBtn: 'Oui',
        cancelBtn: 'Non'
      },
      network: {
        title: 'Réseau',
        wifi: {
          title: 'Wi-Fi',
          description: 'Configurez le Wi-Fi',
          apMode: 'Le mode AP est activé, connectez-vous au Wi-Fi en scannant le code QR',
          connect: 'Connecter le Wi-Fi',
          connectDesc1: 'Veuillez saisir le SSID du réseau et le mot de passe',
          connectDesc2: 'Veuillez saisir le mot de passe pour rejoindre ce réseau',
          disconnect: 'Voulez-vous vraiment déconnecter le réseau ?',
          failed: 'Échec de la connexion, veuillez réessayer.',
          ssid: 'Nom',
          password: 'Mot de passe',
          joinBtn: 'Rejoindre',
          confirmBtn: 'OK',
          cancelBtn: 'Annuler'
        },
        tls: {
          description: 'Activer le protocole HTTPS',
          tip: "Attention : l'utilisation de HTTPS peut augmenter la latence, surtout en mode vidéo MJPEG.",
          restarting: "Redémarrage du serveur de l'appareil, cela prend environ deux minutes...",
          waiting: "En attente d'une nouvelle réponse de l'appareil...",
          waitingHttp: "Retour en http. Rechargez cette page si elle ne s'ouvre pas d'elle-même."
        },
        ethernet: {
          title: 'Adresse IP',
          description: 'Configurez la façon dont NanoKVM obtient son adresse sur le réseau filaire',
          dhcp: 'DHCP',
          manual: 'Manuel',
          networkDetails: 'Détails du réseau',
          interface: 'Interface',
          ipAddress: 'Adresse IP',
          subnetMask: 'Masque de sous-réseau',
          router: 'Routeur',
          save: 'Appliquer',
          invalidAddress: 'Veuillez saisir une adresse IP valide',
          invalidMask:
            'Veuillez saisir un masque de sous-réseau valide, par exemple 255.255.255.0 ou 24',
          invalidRouter: 'Veuillez saisir une adresse de routeur valide',
          addressRequired: 'Une adresse IP est requise',
          maskRequired: 'Un masque de sous-réseau est requis',
          applyTitle: "Modifier l'adresse de NanoKVM ?",
          applyWarning:
            "La connexion à cette page sera perdue. NanoKVM applique la nouvelle adresse et attend {{seconds}} secondes que vous l'atteigniez à cette adresse. L'atteindre conserve la modification. Si rien ne l'atteint, NanoKVM rétablit les paramètres précédents.",
          applyConfirm: 'Appliquer',
          applyCancel: 'Annuler',
          applyFailed: "Échec de l'application de l'adresse",
          trialTitle: 'En attente de confirmation',
          trialDhcp: 'NanoKVM demande une adresse au DHCP.',
          trialStatic: 'NanoKVM est maintenant à {{address}}.',
          trialInstruction:
            "Ouvrez NanoKVM à sa nouvelle adresse et connectez-vous s'il le demande. L'atteindre conserve la modification. Si rien n'atteint NanoKVM dans {{seconds}} secondes, il rétablit les paramètres précédents.",
          trialOpen: 'Ouvrir la nouvelle adresse',
          trialKeep: 'Conserver ces paramètres',
          trialKept: 'La nouvelle adresse est enregistrée',
          trialKeepFailed: 'Échec de la conservation des paramètres',
          trialGone: 'La modification a déjà été annulée. Veuillez réessayer.',
          unsaved: 'Modifications non enregistrées'
        },
        dns: {
          title: 'DNS',
          description: 'Configurer les serveurs DNS pour NanoKVM',
          mode: 'Mode',
          dhcp: 'DHCP',
          manual: 'Manuel',
          add: 'Ajouter un DNS',
          save: 'Enregistrer',
          invalid: 'Veuillez saisir une adresse IP valide',
          noDhcp: "Aucun DNS DHCP n'est actuellement disponible",
          saved: 'Paramètres DNS enregistrés',
          saveFailed: "Échec de l'enregistrement des paramètres DNS",
          unsaved: 'Modifications non enregistrées',
          maxServers: '{{count}} serveurs DNS maximum autorisés',
          dnsServers: 'Serveurs DNS',
          dhcpServersDescription: 'Les serveurs DNS sont obtenus automatiquement par DHCP',
          manualServersDescription: 'Les serveurs DNS peuvent être modifiés manuellement',
          networkDetails: 'Détails du réseau',
          interface: 'Interface',
          ipAddress: 'Adresse IP',
          subnetMask: 'Masque de sous-réseau',
          router: 'Routeur',
          none: 'Aucun'
        }
      },
      vpn: {
        loading: 'Chargement...',
        okBtn: 'Oui',
        cancelBtn: 'Non',
        restart: 'Redémarrer {{name}} ?',
        stop: 'Arrêter {{name}} ?',
        stopDesc:
          "Le démon s'arrête maintenant. Le démarrage automatique est un réglage distinct et reste inchangé.",
        update: 'Mettre à jour {{name}} vers {{version}} ?',
        updateDesc: "Le démon redémarre s'il est en cours d'exécution. La connexion est conservée.",
        notInstall: "{{name}} n'est pas installé.",
        install: 'Installer',
        installing: 'Installation',
        installFailed: "Échec de l'installation",
        retry: 'Réessayer',
        notRunning: "{{name}} n'est pas en cours d'exécution. Démarrez-le pour continuer.",
        run: 'Démarrer',
        boot: 'Démarrage automatique',
        bootDesc: 'Démarrer {{name}} au démarrage du KVM.',
        enable: 'Activer {{name}}',
        control: 'Serveur de contrôle',
        connected: 'Connecté',
        disconnected: 'Non connecté',
        deviceName: "Nom de l'appareil",
        deviceIP: "IP de l'appareil",
        account: 'Compte',
        version: 'Version',
        uptime: 'Temps de fonctionnement',
        peers: 'Pairs',
        noPeers: "Aucun pair pour l'instant.",
        online: 'En ligne',
        offline: 'Hors ligne',
        memory: 'Mémoire',
        daemonRss: 'Démon',
        group: 'Groupe des modules',
        high: 'ralenti au-delà de {{size}}',
        max: 'arrêté par le noyau au-delà de {{size}}',
        noGroup: 'Aucun groupe de mémoire pour les modules sur cette carte.',
        uninstall: 'Désinstaller {{name}}',
        uninstallDesc:
          'Êtes-vous sûr de vouloir désinstaller {{name}} ? La connexion reste enregistrée sur la carte.',
        blocked:
          "{{other}} est en cours d'exécution ou démarre automatiquement. Un seul VPN peut tourner à la fois : arrêtez d'abord {{other}} et désactivez son démarrage automatique.",
        swap: {
          title: "Mémoire d'échange",
          tip: "Si le démon manque de mémoire, essayez d'activer la mémoire d'échange. La taille du fichier d'échange est alors fixée à 256MB par défaut, ce qui peut être modifié dans \"Paramètres > Appareil\"."
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: "Veuillez rafraîchir et réessayer. Ou essayez d'installer manuellement",
        download: 'Télécharger le',
        package: "paquet d'installation",
        unzip: 'et décompressez-le',
        upTailscale: 'Téléverser tailscale dans le répertoire NanoKVM /usr/sbin/',
        upTailscaled: 'Téléverser tailscaled dans le répertoire NanoKVM /usr/sbin/',
        refresh: 'Rafraîchir la page courante',
        notLogin: "L'appareil n'est pas relié. Connectez-vous et liez cet appareil à votre compte.",
        urlPeriod: "L'URL est valide pendant 10 minutes",
        login: 'Connexion',
        loginSuccess: 'Connexion réussie',
        logout: 'Déconnexion',
        logoutDesc: 'Êtes-vous sûr de vouloir vous déconnecter?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          "Cet appareil n'a pas encore rejoint de réseau NetBird. Rejoignez-en un avec une clé de configuration, ou connectez-vous via SSO.",
        setupKey: 'Clé de configuration',
        setupKeyPlaceholder: 'Collez une clé de configuration depuis le tableau de bord NetBird',
        join: 'Rejoindre',
        or: 'ou',
        sso: 'Se connecter via SSO',
        urlPeriod: 'Cette URL est valide pendant 10 minutes',
        loginSuccess: 'Connexion réussie',
        logout: 'Désinscrire',
        logoutDesc:
          'La désinscription retire ce pair de votre compte NetBird et supprime sa configuration ici. Pour rejoindre à nouveau, il faudra une clé de configuration ou une connexion SSO, et le pair pourra recevoir une nouvelle IP. Continuer ?'
      },
      update: {
        title: 'Vérifier les mises à jour',
        queryFailed: 'Impossible de vérifier les mises à jour. Veuillez réessayer.',
        updateFailed: 'Mise à jour échouée. Veuillez réessayer.',
        isLatest: 'Vous avez déjà la dernière version.',
        available: 'Une mise à jour est disponible. Voulez-vous vraiment mettre à jour?',
        updating: 'Mise à jour en cours. Veuillez patienter...',
        confirm: 'Confirmer',
        cancel: 'Annuler',
        preview: 'Aperçu des mises à jour',
        previewDesc:
          "Bénéficiez d'un accès anticipé aux nouvelles fonctionnalités et améliorations",
        previewTip:
          'Veuillez noter que les versions préliminaires peuvent contenir des bugs ou des fonctionnalités incomplètes!',
        customServer: {
          title: 'Serveur de mise à jour personnalisé',
          desc: 'Rechercher et télécharger les mises à jour en ligne depuis un serveur spécifié',
          invalidUrl:
            'Saisissez un répertoire de serveur HTTP ou HTTPS valide, sans paramètres de requête, fragment ni latest.json.',
          loadFailed: 'Impossible de charger la configuration du serveur de mise à jour.',
          saveFailed: 'Impossible d’enregistrer la configuration du serveur de mise à jour.',
          saved: 'Configuration du serveur de mise à jour enregistrée.',
          save: 'Enregistrer',
          confirmTitle: 'Utiliser un serveur de mise à jour personnalisé ?',
          confirmDesc:
            'SHA-512 vérifie uniquement que le paquet correspond au manifeste fourni par ce serveur. Cela ne prouve pas que le paquet est une version officielle de NanoKVM. Un serveur défectueux ou malveillant peut rendre l’appareil inutilisable, entraîner une perte de données ou compromettre le système.',
          confirm: 'Utiliser quand même',
          useSipeed: 'Utiliser le serveur officiel de Sipeed',
          previewDisabled:
            'Les mises à jour en préversion ne sont pas disponibles lorsqu’un serveur de mise à jour personnalisé est activé.'
        },
        offline: {
          title: 'Mises à jour hors ligne',
          desc: "Mise à jour via le package d'installation local",
          upload: 'Téléverser',
          checksumPlaceholder: 'Somme de contrôle SHA-256 (facultative)',
          invalidChecksum: 'La somme de contrôle SHA-256 doit contenir 64 caractères hexadécimaux.',
          checksumMismatch: 'La vérification SHA-256 a échoué. Le paquet est peut-être endommagé.',
          invalidName:
            'Format de nom de fichier invalide. Veuillez télécharger à partir des versions de GitHub.',
          updateFailed: 'Mise à jour échouée. Veuillez réessayer.'
        }
      },
      account: {
        title: 'Compte',
        webAccount: 'Nom du compte Web',
        role: 'Rôle',
        roles: { admin: 'Administrateur', user: 'Utilisateur' },
        password: 'Mot de passe',
        updateBtn: 'Mettre à jour',
        logoutBtn: 'Déconnexion',
        logoutDesc: 'Êtes-vous sûr de vouloir vous déconnecter?',
        okBtn: 'Oui',
        cancelBtn: 'Non',
        users: {
          title: 'Utilisateurs',
          create: 'Créer un utilisateur',
          enabled: 'Activé',
          disabled: 'Désactivé',
          deviceOwner: "Propriétaire de l'appareil",
          resetPassword: 'Réinitialiser le mot de passe',
          delete: 'Supprimer',
          deleteConfirm: 'Supprimer cet utilisateur et révoquer toutes ses sessions ?',
          created: 'Utilisateur créé',
          deleted: 'Utilisateur supprimé',
          passwordUpdated: 'Mot de passe mis à jour',
          loadFailed: 'Impossible de charger les utilisateurs',
          saveFailed: "Impossible d'enregistrer l'utilisateur",
          deleteFailed: "Impossible de supprimer l'utilisateur"
        }
      },
      apiKeys: {
        title: 'Clés API',
        description:
          "Une clé agit au nom de son propriétaire, avec le rôle de cet utilisateur. Envoyez-la en Authorization: Bearer <key> pour les métriques et l'API, ou en X-Auth-Token pour Redfish.",
        name: 'Nom',
        namePlaceholder: 'À quoi sert la clé, par exemple prometheus',
        nameRequired: 'Donnez un nom à la clé',
        nameTooLong: 'Le nom ne doit pas dépasser 64 caractères',
        unnamed: '(sans nom)',
        create: 'Créer une clé',
        created: 'Créée',
        owner: 'Propriétaire',
        empty: 'Aucune clé API',
        newKeyTitle: 'Votre nouvelle clé API',
        newKeyWarning:
          "Copiez la clé maintenant. Elle n'est pas stockée et ne pourra plus être affichée. Si vous la perdez, révoquez-la et créez-en une autre.",
        copy: 'Copier',
        copied: 'Copiée',
        copyFailed: 'Échec de la copie. Copiez manuellement.',
        done: 'Terminé',
        revoke: 'Révoquer',
        revokeConfirmTitle: 'Révoquer cette clé API ?',
        revokeConfirmDesc: 'Tout ce qui utilise "{{name}}" cesse immédiatement de fonctionner.',
        revoked: 'Clé API révoquée',
        loadFailed: 'Impossible de charger les clés API',
        createFailed: 'Impossible de créer la clé API',
        revokeFailed: 'Impossible de révoquer la clé API',
        cancelBtn: 'Annuler'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistante',
      empty: 'Ouvrez le panneau et démarrez une tâche pour commencer.',
      inputPlaceholder: 'Décrivez ce que vous voulez que le PicoClaw fasse',
      newConversation: 'Nouvelle conversation',
      processing: 'Traitement...',
      agent: {
        defaultTitle: 'Assistant général',
        defaultDescription: "Aide générale sur le chat, la recherche et l'espace de travail.",
        kvmTitle: 'Contrôle à distance',
        kvmDescription: "Faites fonctionner l'hôte distant via NanoKVM.",
        switched: "Rôle d'agent modifié",
        switchFailed: "Échec du changement de rôle d'agent"
      },
      send: 'Envoyer',
      cancel: 'Annuler',
      status: {
        connecting: 'Connexion à la passerelle...',
        connected: 'Session PicoClaw connectée',
        disconnected: 'Session PicoClaw fermée',
        stopped: "Demande d'arrêt envoyée",
        runtimeStarted: 'Runtime PicoClaw démarré',
        runtimeStartFailed: 'Échec du démarrage du runtime PicoClaw',
        runtimeStopped: 'Runtime PicoClaw arrêté',
        runtimeStopFailed: "Échec de l'arrêt du runtime PicoClaw",
        controlSwitchedToMCP: 'Contrôle transféré au service MCP externe'
      },
      connection: {
        runtime: {
          checking: 'Vérification',
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime prêt',
          stopped: 'Runtime arrêté',
          blockedByMCP: 'Le contrôle MCP externe est actif',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime indisponible',
          configError: 'Erreur de configuration'
        },
        transport: {
          connecting: 'Connexion',
          connected: 'Connecté',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'Inactif',
          busy: 'Occupé'
        }
      },
      message: {
        toolAction: 'Action',
        observation: 'Observation',
        screenshot: "Capture d'écran"
      },
      overlay: {
        locked: "PicoClaw contrôle l'appareil. La saisie manuelle est suspendue."
      },
      control: {
        picoclaw: "Contrôle de l'appareil : PicoClaw",
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: "Contrôle de l'appareil : MCP externe",
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: "Contrôle de l'appareil : désactivé",
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Accorder le contrôle',
        release: 'Libérer',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'Contrôle PicoClaw accordé',
        released: 'Contrôle PicoClaw libéré',
        grantFailed: "Échec de l'octroi du contrôle PicoClaw",
        releaseFailed: 'Échec de la libération du contrôle PicoClaw',
        grantConfirmTitle: "Basculer le contrôle de l'appareil vers PicoClaw ?",
        grantConfirmDesc: "Les écritures d'appareil du MCP externe seront interrompues."
      },
      install: {
        install: 'Installer PicoClaw',
        installing: 'Installation de PicoClaw',
        success: 'PicoClaw installé avec succès',
        failed: "Échec de l'installation de PicoClaw",
        uninstalling: 'Désinstallation du runtime...',
        uninstalled: 'Runtime désinstallé avec succès.',
        uninstallFailed: 'Échec de la désinstallation.',
        requiredTitle: "PicoClaw n'est pas installé",
        requiredDescription: 'Installez PicoClaw avant de démarrer le runtime PicoClaw.',
        progressDescription: "PicoClaw est en cours de téléchargement et d'installation.",
        stages: {
          preparing: 'Préparation',
          downloading: 'Téléchargement',
          extracting: 'Extraction',
          verifying: 'Vérification',
          installing: 'Installation',
          installed: 'Installé',
          install_timeout: 'Délai expiré',
          install_failed: 'Échec'
        }
      },
      model: {
        requiredTitle: 'La configuration du modèle est requise',
        requiredDescription: "Configurez le modèle PicoClaw avant d'utiliser le chat PicoClaw.",
        docsTitle: 'Guide de configuration',
        docsDesc: 'Modèles et protocoles pris en charge',
        menuLabel: 'Configurer le modèle',
        modelIdentifier: 'Identifiant du modèle',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Clé API',
        apiKeyPlaceholder: 'Saisissez la clé API du modèle',
        save: 'Enregistrer',
        saving: 'Enregistrement',
        saved: 'Configuration du modèle enregistrée',
        saveFailed: "Échec de l'enregistrement de la configuration du modèle",
        invalid: 'L’identifiant du modèle, l’API Base URL et la clé API sont requis'
      },
      uninstall: {
        menuLabel: 'Désinstaller',
        confirmTitle: 'Désinstaller PicoClaw',
        confirmContent:
          "Êtes-vous sûr de vouloir désinstaller PicoClaw? Cela supprimera l'exécutable et tous les fichiers de configuration.",
        confirmOk: 'Désinstaller',
        confirmCancel: 'Annuler'
      },
      history: {
        title: 'Historique',
        loading: 'Chargement des sessions...',
        emptyTitle: "Pas encore d'historique",
        emptyDescription: 'Les sessions précédentes PicoClaw apparaîtront ici.',
        loadFailed: "Échec du chargement de l'historique de la session",
        deleteFailed: 'Échec de la suppression de la session',
        deleteConfirmTitle: 'Supprimer la session',
        deleteConfirmContent: 'Etes-vous sûr de vouloir supprimer « {{title}} »?',
        deleteConfirmOk: 'Supprimer',
        deleteConfirmCancel: 'Annuler',
        messageCount_one: '{{count}} message',
        messageCount_other: '{{count}} messages',
        messageCount: '{{count}} messages'
      },
      config: {
        startRuntime: 'Démarrer PicoClaw',
        stopRuntime: 'Arrêter PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Transférer le contrôle à PicoClaw ?',
        enableConfirmDesc: 'Le démarrage de PicoClaw désactivera le service MCP externe.',
        enableConfirmOk: 'Démarrer PicoClaw',
        enableConfirmCancel: 'Annuler',
        title: 'Démarrer PicoClaw',
        description: "Démarrez le runtime pour commencer à utiliser l'assistant PicoClaw.",
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'Nous avons rencontré un problème',
      refresh: 'Actualiser',
      panel: 'Cette partie de la page a cessé de fonctionner',
      retry: 'Réessayer'
    },
    fullscreen: {
      toggle: 'Basculer vers le plein écran'
    },
    input: {
      disconnected: 'Le clavier et la souris ne sont pas connectés',
      disconnectedTls:
        "Le navigateur a refusé la connexion sécurisée qui transporte le clavier et la souris, et il le fait sans demander. Le certificat généré par cet appareil n'est pas encore approuvé. Ouvrez cette adresse dans un nouvel onglet, acceptez le certificat, puis rechargez. Installer le certificat est la solution fiable.",
      disconnectedNever:
        "La connexion qui transporte le clavier et la souris n'a pas pu être ouverte. Le reste de la page fonctionne car il ne l'utilise pas. Vérifiez que rien entre vous et l'appareil ne la bloque.",
      disconnectedDropped:
        "La connexion qui transporte le clavier et la souris a été perdue et n'est pas revenue. Elle se rétablit d'elle-même après un redémarrage ; si le problème persiste, rechargez la page.",
      hidDisabled: 'Le HID est désactivé sur cet appareil (/boot/disable_hid).',
      keyFailed: "La touche n'a pas pu être envoyée."
    },
    speaker: { title: 'Haut-parleur', unmute: 'Réactiver le son', mute: 'Couper le son' },
    menu: {
      collapse: 'Réduire le menu',
      expand: 'Développer le menu'
    },
    ion: {
      checking: 'Vérification de la mémoire vidéo avant de lancer le flux...',
      warn: "La mémoire vidéo est faible. Un seul redémarrage du serveur l'épuiserait. Redémarrez quand cela vous convient.",
      criticalTitle: 'Mémoire vidéo insuffisante pour lancer le flux',
      criticalBody:
        "Lancer la vidéo épuiserait la mémoire réservée et arrêterait le serveur. Toutes les autres fonctions marchent encore, y compris le contrôle de l'alimentation et le redémarrage. Seul un redémarrage de NanoKVM récupère cette mémoire.",
      criticalContinue: 'Lancer la vidéo quand même',
      criticalReboot: 'Redémarrer NanoKVM',
      criticalRebooting: 'Redémarrage...'
    }
  }
};

export default fr;
