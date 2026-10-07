const tr = {
  translation: {
    feedback: {
      enabled: '{{name}} etkinleştirildi',
      disabled: '{{name}} devre dışı bırakıldı',
      failed: 'İstek başarısız oldu. Yeniden deneyin.',
      network: 'Cihaza ulaşılamadı. Bağlantıyı kontrol edip yeniden deneyin.',
      saved: 'Kaydedildi',
      timeout: 'Cihazın yanıtı çok uzun sürdü. Yeniden deneyin.'
    },
    common: {
      copy: 'Kopyala',
      copied: 'Kopyalandı',
      copyFailed: 'Kopyalanamadı. Metni seçip elle kopyalayın.',
      notUpdating: 'Güncellenmiyor: son yenileme başarısız oldu.',
      off: 'Kapalı',
      running: 'Çalışıyor',
      save: 'Kaydet',
      cancel: 'İptal',
      delete: 'Sil',
      remove: 'Kaldır'
    },
    head: {
      desktop: 'Uzak masaüstü',
      login: 'Giriş',
      changePassword: 'Şifreyi değiştir',
      terminal: 'Uçbirim',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Parola değiştirildi. Yeni parolayla oturum açın.',
      cookieRejected:
        'Tarayıcı oturumu kaydetmeyi reddetti. Önceki bir HTTPS oturumundan kalan çerez, şifrelenmemiş http üzerinden değiştirilemez. Bu adres için çerezleri temizleyin veya gizli bir pencere açın ve yeniden giriş yapın.',
      login: 'Giriş',
      placeholderUsername: 'Kullanıcı Adı',
      placeholderPassword: 'Şifre',
      placeholderCurrentPassword: 'Mevcut şifre',
      placeholderPassword2: 'Şifrenizi tekrar deneyiniz',
      noEmptyUsername: 'Kullanıcı adı gereklidir',
      noEmptyPassword: 'Şifre gereklidir',
      passwordLength: 'Şifre 8 ile 72 karakter arasında olmalıdır',
      noAccount:
        'Kullanıcı verileri alınırken hata yaşandı, lütfen sayfayı yenileyiniz ya da şifrenizi sıfırlayınız',
      invalidUser: 'Yanlış kullanıcı adı ya da şifre',
      locked: 'Çok fazla giriş yapıldı, lütfen daha sonra tekrar deneyin',
      globalLocked: 'Sistem koruma altında, lütfen daha sonra tekrar deneyin',
      error: 'Beklenmedik bir hata',
      invalidCurrentPassword: 'Mevcut şifre yanlış',
      changePassword: 'Şifrenizi değiştiriniz',
      changePasswordDesc: 'Güvenlik sebebiyle lütfen şifrenizi değiştiriniz!',
      differentPassword: 'Şifreler eşleşmemektedir',
      illegalUsername: 'Kullanıcı adı istenmeyen karakterler içermektedir',
      illegalPassword: 'Şifre istenmeyen karakterler içermektedir',
      forgetPassword: 'Şifremi unuttum',
      ok: 'Tamam',
      cancel: 'İptal',
      loginButtonText: 'Giriş',
      tips: {
        reset1:
          'Şifreleri sıfırlamak için IronKVM üzerinde bulunan BOOT tuşuna 10 saniye boyunca basılı tutun.',
        reset3: 'Arayüz varsayılan hesap:',
        reset4: 'Güvenli Kabuk Bağlantısı (SSH) varsayılan hesap:',
        change1: 'Bu işlem şu şifreleri değiştiricektir:',
        change2: 'Arayüz giriş şifresi',
        change3: 'Sistem yöneticisi şifresi (Güvenli Kabuk Bağlantısı (SSH) giriş şifresi)',
        change4:
          'Şifreleri sıfırlamak için IronKVM üzerinde bulunan BOOT tuşuna 10 saniye boyunca basılı tutun.',
        resetDocs: 'Ayrıntılı adımlar için donanım belgelerine bakın:',
        hardwareDocs: 'Sipeed NanoKVM wiki'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'IronKVM için Wi-Fi ayarlarını ayarlayın',
      success: "IronKVM'in bağlantı durumunu kontrol edin ve yeni IP adresini ziyaret edin.",
      failed: 'İşlem başarısız oldu, lütfen tekrar deneyiniz.',
      invalidMode:
        'Geçerli mod ağ kurulumunu desteklemiyor. Lütfen cihazınıza gidin ve Wi-Fi yapılandırma modunu etkinleştirin.',
      confirmBtn: 'Tamam',
      finishBtn: 'Bitti',
      ap: {
        authTitle: 'Kimlik Doğrulaması Gerekli',
        authDescription: 'Devam etmek için lütfen AP şifresini girin',
        authFailed: 'Geçersiz AP şifresi',
        passPlaceholder: 'AP şifre',
        verifyBtn: 'Doğrula'
      },
      ssidRequired: 'Ağ adını girin, en fazla 32 karakter',
      passwordLength: 'Parola 8 ile 63 karakter arasındadır. Açık ağ için boş bırakın.',
      passwordOptional: 'Parola (açık ağ için boş)',
      lost: 'Kart yanıt vermeyi bıraktı. Ağa katılıp kurulum erişim noktasını kapatmış olabilir. Erişim noktası geri gelirse katılma başarısız olmuştur: ona yeniden bağlanıp tekrar deneyin.',
      done: 'Kurulum tamamlandı. Bu cihazı her zamanki ağınıza yeniden bağlayın ve kartı yeni adresinden açın.'
    },
    screen: {
      viewOnly: 'Yalnızca izle',
      viewOnlyTip:
        'Bu sekme ana bilgisayara klavye ve fare girdisi göndermeyi bırakır. Betikler, fare titreştirici ve diğer izleyiciler etkilenmez.',
      viewOnlyOff: 'Yalnızca izlemeyi kapat',
      viewOnlyBlocked: 'Yalnızca izle açık, ana bilgisayara hiçbir şey gönderilmedi',
      pauseHidden: 'Sekme gizliyken duraklat',
      pauseHiddenTip:
        'Bu sekme gizlendikten birkaç saniye sonra görüntüyü ve sesi durdurur, döndüğünüzde yeniden başlatır.',
      screenshot: 'Ekran görüntüsü',
      screenshotTip: 'Ana bilgisayar ekranını tam yakalama boyutunda PNG olarak kaydeder.',
      screenshotFailed: 'Ekran görüntüsü alınamadı',
      stream: {
        ok: 'görüntü tamam',
        noSignal: 'sinyal yok',
        failed: 'akış başarısız'
      },
      codecNoWebrtcHevc: 'Bu tarayıcı WebRTC üzerinden H.265 alamıyor',
      codecNoHevc: 'Bu tarayıcı H.265 çözemiyor',
      codecNote:
        'Kartta tek bir kodlayıcı var, bu yüzden bu tüm izleyicilerin yayınını değiştirir.',
      codec: 'Kodek',
      updateFailed: 'Ayar uygulanmadı',
      scale: 'Ölçek',
      title: 'Ekran',
      video: 'Görüntü modu',
      videoDirectTips: 'kullanmak için "Ayarlar > Cihaz" HTTPS aktif edin',
      resolution: 'Çözünürlük',
      aspect: 'En boy oranı',
      aspectKeep: 'Otomatik (kaynağın oranını koru)',
      aspectStretch: 'Çözünürlüğe uzat',
      aspectTips:
        'Otomatik, ana bilgisayar ekranının şeklini seçilen yükseklikte korur: 1920x1200 bir ekran 1728x1080 olarak gönderilir. Uzat, seçilen çözünürlüğü doldurur ve 16:9 olmayan ekranları bozar.',
      ocr: {
        title: 'Metni Oku (OCR)',
        tips: 'Metin bu tarayıcıda tanınır. Kopyalamadan önce düzeltebilirsiniz.',
        hint: 'Okunacak metnin üzerinde sürükleyin. İptal etmek için Esc tuşuna basın.',
        noPicture: 'Görüntüyü bekleyin, ardından okunacak metnin üzerinde sürükleyin.',
        cancel: 'İptal',
        language: 'Dil',
        languages: {
          eng: 'İngilizce'
        },
        preview: 'Seçili alan',
        capturing: 'Ekran yakalanıyor...',
        loading: 'Metin tanıma yükleniyor...',
        recognizing: 'Metin okunuyor...',
        noText: 'Seçili alanda metin bulunamadı.',
        copy: 'Kopyala',
        copied: 'Panoya kopyalandı',
        copyFailed: 'Panoya kopyalanamadı',
        selectAgain: 'Yeniden Seç',
        unsupported:
          'Bu tarayıcı metin tanıma çalıştıramıyor. Güncel tarayıcılarda bulunan WebAssembly SIMD desteği gerekir.',
        captureFailed: 'Ekran yakalanamadı.',
        outside: 'Seçili alan görüntünün dışında.',
        recognizeFailed: 'Metin tanıma başarısız oldu.'
      },
      controlRegion: {
        title: 'Fare Kalibrasyonu',
        description:
          'Kontrol edilen cihaz 16:9 dışında bir çözünürlük kullandığında ve imleç yatay veya dikey olarak hizalanmadığında bu ayarı kullanın.',
        off: 'Kapalı',
        auto: 'Otomatik',
        autoWarning:
          'Kullanıcı uygulamasının arka planı tamamen siyah olduğunda kalibrasyon başarısız olabilir.',
        manual: 'Manuel',
        selectedResolution: 'Seçili Alan Çözünürlüğü',
        unused: 'Kullanılmıyor',
        originalResolution: 'Orijinal Çözünürlük',
        selectResolution: 'Orijinal çözünürlüğü seçin',
        addResolution: 'Özel çözünürlük ekle',
        add: 'Ekle',
        duplicateResolution: 'Bu çözünürlük zaten mevcut.',
        width: 'Genişlik',
        height: 'Yükseklik',
        apply: 'Hesapla ve Uygula',
        invalidResolution: 'Video hazır olduktan sonra geçerli bir orijinal çözünürlük girin.',
        select: 'Alan Seç',
        clear: 'Otomatik Ayarı Geri Yükle',
        saveFailed: 'Giriş alanı kaydedilemedi.',
        tooSmall: 'Seçili alan çok küçük.',
        previewUnavailable: 'Önizleme kullanılamıyor',
        clearConfirm: 'Otomatik siyah kenar algılama geri yüklensin mi?',
        dragHint: 'Uzak masaüstü alanını seçmek için sürükleyin',
        finish: 'Bitti',
        confirm: 'Onayla',
        cancel: 'İptal'
      },
      auto: 'Otomatik',
      autoTips:
        'Belirli çözünürlüklerde ekran yırtılması veya fare kayması meydana gelebilir. Bu durumda uzak ana bilgisayarın çözünürlüğünü ayarlamayı ya da otomatik modu devre dışı bırakmayı deneyin.',
      fps: 'Saniyedeki kare sayısı',
      customizeFps: 'Kişiselleştir',
      quality: 'Kalite',
      qualityLossless: 'En iyi',
      qualityHigh: 'Yüksek',
      qualityMedium: 'Orta',
      qualityLow: 'Düşük',
      frameDetect: 'Kareleri algıla',
      frameDetectTip:
        'Gönderilen kareler arasındaki farkı hesaplar. Uzak ana bilgisayardan gönderilen yayında bir değişiklik yoksa görüntü yayınını durdurur.',
      resetHdmi: 'HDMI sıfırla',
      mixedH264: {
        title: 'H.264 akış çakışması',
        description:
          'H.264 Direct ve H.264 WebRTC aynı anda kullanılıyor. Bu, ekran yırtılmasına veya bozuk videoya neden olabilir. Lütfen yalnızca bir H.264 modu kullanın.'
      },
      webrtcConnectionFailed: {
        title: 'WebRTC bağlantısı başarısız',
        description: 'Ağ bağlantısını kontrol edin veya video modunu değiştirin.'
      },
      captureStatus: {
        hdmiError: 'HDMI ekran hatası',
        unsupportedResolution: 'Geçerli çözünürlük desteklenmiyor',
        retrieving: 'Ekran alınıyor...',
        changingResolution: 'Çözünürlük değiştiriliyor...',
        updateFailed: 'Ekran şu anda güncellenemiyor',
        videoError: 'Video görüntüleme hatası',
        noHdmi: 'HDMI sinyali algılanmadı',
        unavailable: 'Ekran şu anda gösterilemiyor'
      },
      directConnectionFailed: 'Video akışı bağlantısı başarısız'
    },
    keyboard: {
      close: 'Kapat',
      title: 'Klavye',
      paste: 'Yapıştır',
      tips: 'Metni ana makinede tuş basışları olarak yazar. Ana makinenin kullandığı klavye düzenini seçin.',
      placeholder: 'Girdi',
      submit: 'Gönder',
      virtual: 'Klavye',
      readClipboard: 'Panodan Oku',
      clipboardPermissionDenied:
        'Pano izni reddedildi. Lütfen tarayıcınızda pano erişimine izin verin.',
      clipboardReadError: 'Pano okunamadı',
      mediaKeys: {
        title: 'Medya tuşları',
        mute: 'Sessiz',
        volumeDown: 'Sesi azalt',
        volumeUp: 'Sesi artır',
        previous: 'Önceki parça',
        playPause: 'Oynat veya duraklat',
        next: 'Sonraki parça',
        stop: 'Durdur'
      },
      pasting: {
        layout: 'Ana makinedeki klavye düzeni',
        layouts: {
          us: 'İngilizce (ABD)',
          uk: 'İngilizce (Birleşik Krallık)',
          de: 'Almanca',
          fr: 'Fransızca',
          es: 'İspanyolca',
          it: 'İtalyanca',
          ptBr: 'Portekizce (Brezilya)',
          se: 'İsveççe / Fince',
          ru: 'Rusça',
          ja: 'Japonca',
          ko: 'Korece'
        },
        speed: 'Yazma hızı',
        speeds: {
          fast: 'Hızlı',
          normal: 'Normal',
          slow: 'Yavaş'
        },
        estimate: 'Yazma süresi: yaklaşık {{duration}}',
        untypeable: 'Bu düzenin yazamadığı karakterler: {{count}}',
        untypeableAt: 'satır {{line}}, sütun {{column}}',
        skipUntypeable: 'Kalanını yaz',
        shortcut: '{{shortcut}} panodaki metni ana makineye hemen yazar.',
        clipboardUnavailable:
          'Tarayıcı bir sayfanın panoyu yalnızca HTTPS üzerinden okumasına izin verir. Metni Ctrl+V ile kutuya yapıştırın.',
        clipboardEmpty: 'Panoda metin yok.',
        tooLong: 'Metin çok uzun. Sınır {{max}} karakter.',
        inProgress: 'Zaten bir yapıştırma yazılıyor.',
        typing: 'Ana makinede yazılıyor',
        done: 'Metin yazıldı',
        canceled: 'Yapıştırma iptal edildi',
        failed: 'Yapıştırma başarısız oldu',
        cancel: 'İptal',
        controlBusy: 'Klavyeyi başka bir denetleyici kullanıyor.',
        hidError: 'Tuş basışları ana makineye gönderilemedi.'
      },
      shortcut: {
        sendFailed: 'Gönderilmedi: giriş bağlantısı kopuk',
        title: 'Kısayollar',
        custom: 'Özel',
        capture: 'Kısayolu yakalamak için burayı tıklayın',
        clear: 'Temizle',
        save: 'Kaydet',
        captureTips:
          'Windows tuşu gibi sistem düzeyi tuşları yakalamak için tam ekran izni gerekir.',
        enterFullScreen: 'Tam ekran moduna geçiş yapın.'
      },
      leaderKey: {
        saveFailed: 'Lider tuş kaydedilemedi',
        title: 'Leader Tuşu',
        desc: 'Tarayıcı kısıtlamalarını atlayın ve sistem kısayollarını doğrudan uzak ana bilgisayara gönderin.',
        howToUse: 'Nasıl Kullanılır',
        simultaneous: {
          title: 'Eşzamanlı Mod',
          desc1: 'Leader tuşunu basılı tutun, ardından kısayola basın.',
          desc2: 'Sezgisel, ancak sistem kısayollarıyla çakışabilir.'
        },
        sequential: {
          title: 'Sıralı Mod',
          desc1: 'Leader tuşuna basın → kısayola sırayla basın → Leader tuşuna tekrar basın.',
          desc2: 'Daha fazla adım gerektirir ancak sistem çakışmalarını tamamen önler.'
        },
        enable: 'Leader tuşunu etkinleştir',
        tip: 'Leader tuşu olarak atandığında bu tuş yalnızca kısayol tetikleyici olarak çalışır ve varsayılan davranışını kaybeder.',
        placeholder: 'Leader tuşuna basın',
        shiftRight: 'Sağ Shift',
        ctrlRight: 'Sağ Ctrl',
        metaRight: 'Sağ Win',
        submit: 'Gönder',
        recorder: {
          rec: 'KAYIT',
          activate: 'Tuşları etkinleştir',
          input: 'Lütfen kısayola basın...'
        }
      }
    },
    mouse: {
      jiggler: 'Fare kıpırdatıcı',
      keyJiggler: 'Tuş basıcı',
      keyJigglerF15: 'F15 tuşu',
      keyJigglerShift: 'Shift tuşu',
      keyJigglerCtrl: 'Ctrl tuşu',
      keyJigglerF15Tip: 'F15 en az rahatsız edendir: yaygın hiçbir sistem veya uygulama onu kullanmaz',
      title: 'Fare',
      cursor: 'İmleç sitili',
      default: 'Varsayılan imleç',
      pointer: 'Nokta imleç',
      cell: 'Artı imleç',
      text: 'Yazı imleç',
      grab: 'El imleç',
      hide: 'İmleci gizle',
      mode: 'Fare modu',
      absolute: 'Mutlak fare modu',
      relative: 'Bağıl fare modu',
      absoluteShort: 'Mutlak',
      relativeShort: 'Bağıl',
      touch: 'Dokunmatik mod',
      touchShort: 'Dokunmatik',
      absoluteStalled: 'Hedef cihaz mutlak fareyi yok sayıyor',
      absoluteStalledDesc:
        "Hedef cihaz mutlak fare raporlarını almayı bıraktı, bu yüzden imleç hareketleri kayboluyor. Klavye bundan etkilenmez. USB'yi kurtarmak genellikle sorunu giderir; bağıl mod farklı bir uç nokta kullanır.",
      useRelative: 'Bağıl moda geç',
      direction: 'Kaydırma tekerleği yönü',
      scrollUp: 'Bu bilgisayardaki gibi',
      scrollDown: 'Ters (doğal kaydırma)',
      speed: 'Kaydırma tekerleği hızı',
      fast: 'Hızlı',
      slow: 'Yavaş',
      requestPointer: 'Bağıl fare modu kullanılıyor. Masaüstüne tıklayarak imleç elde edinin.',
      resetHid: 'HID’yi sıfırla',
      hidOnly: {
        switchFailed: 'Mod değiştirilemedi. Bağlantıyı kontrol edip yeniden deneyin.',
        title: 'Yalnızca HID modu',
        desc: 'Fare ve klavye yanıt vermeyi durdurursa ve HID sıfırlama yardımcı olmazsa, IronKVM ile cihaz arasında bir uyumluluk sorunu olabilir. Daha iyi uyumluluk için yalnızca HID modunu etkinleştirmeyi deneyin.',
        tip1: 'Yalnızca HID modunu etkinleştirmek sanal U-disk’i ve sanal ağı ayırır',
        tip2: 'Yalnızca HID modunda imaj bağlama devre dışıdır',
        rebuild: 'Mod değiştirmek USB bağlantısını yeniden kurar. IronKVM yeniden başlamaz',
        enable: 'Yalnızca HID modunu etkinleştir',
        disable: 'Yalnızca HID modunu devre dışı bırak'
      },
      resetHidDone: 'USB HID sıfırlandı',
      resetHidFailed: 'USB HID sıfırlanamadı'
    },
    image: {
      driveLoaded: 'imaj takılı',
      driveWarning: 'uyarılara bakın',
      warning: {
        missing: 'İmaj dosyası silindi. Siz çıkarana kadar ana bilgisayar eski kopyayı okur.',
        writable: 'Okuma-yazma: ana bilgisayar bu imajı değiştirebilir.',
        tooBigForCd: 'CD sürücüsü için çok büyük ({{size}}, sınır {{max}}). Diski kullanın.',
        tooSmallForCd: 'CD sürücüsü için çok küçük ({{size}}). Diski kullanın.',
        empty: 'Dosya boş, büyük olasılıkla başarısız bir yükleme veya indirme yüzünden.'
      },
      delete: 'Sil',
      inUse: 'Kullanımda. Silmeden önce çıkarın.',
      retry: 'Yeniden dene',
      loadFailed: 'İmaj listesi yüklenemedi',
      readOnlyLocked: 'Değiştirmek için diski çıkarın. Bir imaj takıldığında geçerli olur.',
      title: 'Disk İmajları',
      loading: 'Yükleniyor...',
      empty: 'Hiçbir şey bulunamadı',
      mountMode: 'Montaj modu',
      mountFailed: 'Bağlantı başarısız oldu',
      mountDesc:
        'Bazı sistemlerde, disk imajını bağlamadan önce uzak ana bilgisayardaki sanal diski çıkarmak gerekir.',
      unmountFailed: 'Bağlantıyı kesme işlemi başarısız oldu',
      unmountDesc:
        'Bazı sistemlerde, görüntünün bağlantısını kesmeden önce uzak ana bilgisayardan manuel olarak çıkarmanız gerekir.',
      refresh: 'Disk imajı listesini yenile',
      disk: 'Disk',
      cdrom: 'CD',
      driveEmpty: 'Boş',
      eject: 'Çıkar',
      readOnly: 'Salt okunur',
      readOnlyTip: 'Diske takılacak bir sonraki imaj için geçerlidir.',
      noDrives: "Sanal sürücü yok. Ayarlar'dan sanal diski açın.",
      insertFailed: 'Takma başarısız oldu',
      ejectFailed: 'Çıkarma başarısız oldu',
      insertInto: 'Sürücüye tak: {{drive}}. Değiştirmek için tıklayın.',
      loadedIn: 'Sürücüde: {{drive}}',
      attention: 'Dikkat',
      deleteConfirm: 'Bu resmi silmek istediğinizden emin misiniz?',
      okBtn: 'Evet',
      cancelBtn: 'Hayır',
      deleteFailed: 'Silme başarısız',
      ventoy: {
        statusNoKernel: 'Bu yazılım desteklemiyor',
        statusNotInstalled: 'Yüklü değil',
        statusReady: 'Hazır',
        statusSelected: 'Seçilen imajlar: {{count}}',
        statusInDrive: 'Disk sürücüsünde, {{size}}',
        noKernel:
          'Bu yazılımın çekirdeğinde device-mapper desteği yok, bu yüzden bu desteğe sahip bir imaj yüklenene kadar Ventoy kullanılamaz.',
        installDesc:
          'Ana bilgisayarı tek bir diskteki birden çok imajdan, onları kopyalamadan başlatın.',
        install: 'Yükle',
        installing: 'Ventoy indiriliyor, yaklaşık 20 MB. Bu birkaç dakika sürebilir.',
        needsData: 'Ventoy, /data bölümü bağlı bir IronKVM imajı gerektirir.',
        uninstall: 'Kaldır',
        uninstallConfirm: 'Ventoy dosyaları kaldırılsın mı?',
        noImages: 'Ventoy diskine konacak imaj yok.',
        onDisk: 'Ventoy diskinde',
        missing: 'Eksik: {{file}}',
        remove: 'Ventoy diskinden çıkar',
        setHint: 'İmaj seçimi yalnızca Ventoy diski hiçbir sürücüde değilken değiştirilebilir.',
        useAsDisk: 'Sanal disk olarak kullan',
        failed: 'Ventoy isteği başarısız',
        secureBoot:
          'Secure Boot açıksa ana bilgisayar, Ventoy anahtarını MokManager içinde bir kez kaydetmelidir. ENROLL_THIS_KEY_IN_MOKMANAGER.cer anahtar dosyası VTOYEFI bölümündedir.',
        readOnly:
          'Ana bilgisayar diski salt okunur görür, bu yüzden Ventoy kalıcılığı ve sürücüdeki ventoy.json çalışmaz.'
      },
      tips: {
        title: 'Nasıl yüklenir',
        usb1: "IronKVM'i bilgisayarınıza USB ile bağlayın.",
        usb2: 'Sanal diskin bağlı olduğundan emin olun (Ayarlar - Sanal Disk).',
        usb3: 'Sanal diski bilgisayarınızda açın ve disk imajı dosyanızı sanal diskin kök dizinine kopyalayın.',
        scp1: 'IronKVM ve bilgisayarınızın aynı yerel ağda bulunduğundan emin olun.',
        scp2: "Bilgisayarınızda uçbirimi açın ve disk imajı dosyanını SCP komudunu kullanarak IronKVM'in /data dizinine yükleyin.",
        scp3: 'Örnek: scp senin-disk-imajı-dizinin root@senin-nanokvm-ip:/data',
        tfCard: 'micro SD kart',
        tf1: 'Bu yöntem Linux sistemlerde desteklenmektedir.',
        tf2: "IronKVM'den micro SD kartı çıkartın(TAM sürüm için öncelikle kutuyu sökün).",
        tf3: 'micro SD kartı kart okuyucusuna takın ve bilgisayarınıza bağlayın.',
        tf4: 'Disk imajı dosyanını micro SD kartın /data dizinine kopyalayın.',
        tf5: "micro SD kartı IronKVM'e geri yerleştirin."
      }
    },
    script: {
      title: 'Betikler',
      upload: 'Yükle',
      run: 'Çalıştır',
      runBackground: 'Arka planda çalıştır',
      runFailed: 'Çalıştırma başarısız oldu',
      attention: 'Dikkat',
      delDesc: 'Bu dosyayı silmek istediğinden emin misin?',
      confirm: 'Evet',
      cancel: 'Hayır',
      delete: 'Sil',
      close: 'Kapat',
      empty: 'Henüz betik yok. Kartta çalıştırmak için bir .sh veya .py dosyası yükleyin.',
      loadFailed: 'Betikler yüklenemedi',
      uploaded: 'Betik yüklendi',
      uploadFailed: 'Betik yüklenemedi',
      started: 'Betik arka planda başlatıldı',
      deleteFailed: 'Betik silinemedi',
      waitLimit: 'Betiğin bitmesi en fazla {{minutes}} dakika bekleniyor.',
      timedOut:
        'Betik {{minutes}} dakikadan uzun sürdü ve bu sayfa beklemeyi bıraktı. Kartta hâlâ çalışıyor olabilir.'
    },
    terminal: {
      invalidBaud: 'Bu baud hızı desteklenmiyor.',
      invalidPort: '/dev altında bir aygıt yolu girin, örneğin /dev/ttyS1.',
      invalidSettings: 'Seri port ayarları geçersiz. Bu, kartın kendi kabuğu.',
      disconnected: "Bağlantı kesildi. Yeniden bağlanmak için Enter'a basın.",
      title: 'Uçbirim',
      nanokvm: 'IronKVM Uçbirimi',
      serial: 'Serial Port Uçbirimi',
      serialPort: 'Seri port',
      serialPortPlaceholder: 'Lütfen serial portunu giriniz',
      baudrate: 'Baud hızı',
      parity: 'Eşlik kontrolü',
      parityNone: 'Yok',
      parityEven: 'Çift',
      parityOdd: 'Tek',
      flowControl: 'Akış kontrolü',
      flowControlNone: 'Yok',
      flowControlSoft: 'Yazılımsal',
      flowControlHard: 'Donanımsal',
      dataBits: 'Veri bitleri',
      stopBits: 'Dur bitleri',
      confirm: 'Tamam'
    },
    wol: {
      no: 'Hayır',
      yes: 'Evet',
      deleteConfirm: 'Bu kayıtlı adres silinsin mi?',
      delete: 'Sil',
      wake: 'Uyandır',
      rename: 'Yeniden adlandır',
      showMac: 'MAC adresini göster',
      showName: 'Adı göster',
      requestFailed: 'Komutu göndermek için cihaza ulaşılamadı',
      deleteFailed: 'Silinemedi',
      renameFailed: 'Yeniden adlandırılamadı',
      title: 'Ağ Üzerinden Uyandırma (WOL)',
      sending: 'Komut gönderiliyor...',
      sent: 'Komut gönderildi',
      input: 'MAC adresi girin',
      ok: 'Tamam'
    },
    download: {
      uploadFailed: 'Yükleme başarısız',
      uploadSuccess: 'Yükleme tamamlandı',
      uploading: 'Yükleniyor: {{file}}',
      downloadingPercent: 'İndiriliyor ({{percent}}): {{file}}',
      downloading: 'İndiriliyor: {{file}}',
      title: 'Disk İmajı İndirici',
      input: 'Uzak imaj URL’sini girin',
      ok: 'Tamam',
      disabled: '/data bölüntüsü salt okunur modda, disk imajı indirilemiyor.',
      uploadbox: 'Dosyayı buraya bırakın veya seçmek için tıklayın',
      inputfile: 'Lütfen resim dosyasını giriniz',
      NoISO: 'ISO yok',
      sha256: 'SHA-256 (isteğe bağlı)',
      sha256Placeholder: '64 karakterlik SHA-256 sağlama toplamını girin',
      invalidSHA256: 'SHA-256, 64 karakterlik bir onaltılık dize olmalıdır',
      failed: 'İndirme başarısız',
      success: 'İndirme başarılı',
      checksumFailed: 'İndirme başarısız: SHA-256 doğrulaması başarısız',
      cancel: 'İptal',
      cancelFailed: 'İndirme iptal edilemedi',
      bootMenu: 'Önyükleme menüsü (netboot.xyz)',
      bootMenuPresent: '{{file}} doğru sağlama toplamıyla zaten cihazda',
      bootMenuDesc: "Sanal CD için netboot.xyz ISO'sunu sağlama toplamı doğrulanmış olarak indirin"
    },
    alerts: {
      title: 'Dikkat gerekiyor',
      temperature: {
        warning: 'Kart {{celsius}} °C. Havanın ona ulaşabildiğini kontrol edin.',
        critical: 'Kart {{celsius}} °C, bu çok sıcak. Havalandırın veya kapatın.'
      },
      storage: {
        warning:
          '{{path}} üzerinde {{total}} alanın yalnızca {{available}} kadarı boş. Büyük imajlar sığmayabilir.',
        critical:
          '{{path}} üzerinde yalnızca {{available}} boş. Yüklemeler, indirmeler ve eklenti kurulumları başarısız olacak. Gerekmeyen imajları silin.'
      },
      vpn: '{{name}} açılışta başlayacak şekilde ayarlı ama çalışmıyor, bu yüzden onun üzerinden uzaktan erişim kesik.',
      openVpn: 'VPN ayarlarını aç',
      stream:
        'Video akışı başarısız oldu. Ekran menüsünde başka bir video modu deneyin veya sayfayı yenileyin.'
    },
    power: {
      resetDesc: 'Ana bilgisayarı hemen yeniden başlatır. Kaydedilmemiş iş kaybolur.',
      powerShortDesc: 'Ana bilgisayarı açar veya işletim sisteminden kapanmasını ister (ACPI).',
      powerLongDesc: 'Ana bilgisayarı düzgün kapatmadan zorla kapatır.',
      hddLed: 'Disk LED',
      hddActive: 'Etkin',
      hddIdle: 'Boşta',
      title: 'Güç',
      showConfirm: 'Doğrulama',
      showConfirmTip: 'Kısa güç basışından önce sor. Sıfırlama ve uzun basış her zaman sorar.',
      reset: 'Sıfırla',
      power: 'Güç',
      powerShort: 'Güç tuşu (bas-çek)',
      powerLong: 'Güç tuşu (uzun bas)',
      resetConfirm: 'Sıfırlama işlemine devam etmek istediğinizden emin misiniz?',
      powerConfirm: 'Güç işlemine devam etmek istediğinizden emin misiniz?',
      okBtn: 'Evet',
      cancelBtn: 'Hayır',
      hostOs: 'Ana makine işletim sistemi',
      hostOsTip: 'USB tuşları olarak gönderilir. Ne yapacaklarına ana makine karar verir.',
      sleep: 'Uyku',
      wake: 'Uyandır',
      wakeKey: 'Shift ile uyandır',
      powerDown: 'Kapat',
      sleepConfirm: 'Ana makine uyku moduna alınsın mı?',
      powerDownConfirm: 'Kapatma tuşu ana makineye gönderilsin mi?',
      wakeTip:
        'Uykudaki bir ana makine, onu uyutan cihazdan gelen Uyandır komutunu çoğu zaman yok sayar. Shift ile uyandır klavyede bir tuşa basar ve bunu daha fazla ana makine kabul eder.',
      led: "Güç LED'i",
      ledOn: 'Yanıyor',
      ledOff: 'Sönük',
      ledUnknown: 'Bilinmiyor',
      ledConnected: "Güç LED'i bağlı",
      ledConnectedTip:
        "Yalnızca ana makinenin güç LED'i konnektörü karta bağlıysa açın. Bu bağlantı olmadan güç durumu bilinemez.",
      ledConnectedFailed: "Güç LED'i ayarı kaydedilemedi",
      powerLongConfirm:
        'Güç düğmesi {{seconds}} sn basılı tutulsun mu? Bu, kapatma yapmadan gücü keser.',
      done: 'Düğmeye basıldı',
      failed: 'Düğmeye basılamadı'
    },
    settings: {
      title: 'Ayarlar',
      nav: {
        system: 'Sistem',
        network: 'Ağ',
        access: 'Erişim',
        integrations: 'Entegrasyonlar',
        boot: 'Önyükleme ve medya',
        browser: 'Bu tarayıcı',
        search: 'Ayar bul',
        noMatch: 'Eşleşen ayar yok',
        locked: 'Bir işlem sürüyor. Bitene kadar diğer sayfalar ve kapatma kullanılamaz.',
        vpnProvider: 'VPN sağlayıcısı'
      },
      mcp: {
        keyNote:
          'MCP aşağıda gösterilen kendi API anahtarını kullanır. API Anahtarları sayfasındaki anahtarlar burada çalışmaz.',
        title: 'MCP Hizmeti',
        service: 'MCP uzaktan kumanda',
        serviceDesc:
          'Güvenilir MCP istemcilerinin klavye ve fareyi kontrol etmesine ve ekran görüntüsü almasına izin verin',
        securityWarning:
          'Bu API anahtarına sahip herkes uzak ana bilgisayarı kontrol edebilir ve ekranını görebilir. HTTPS kullanın ve hizmeti yalnızca güvenilir ağlarda etkinleştirin.',
        endpoint: 'Uç nokta',
        apiKey: 'API anahtarı',
        regenerateConfirmTitle: 'MCP API anahtarı yeniden oluşturulsun mu?',
        regenerateConfirmDesc: 'Geçerli anahtar hemen çalışmayı durduracaktır.',
        enableConfirmTitle: 'Harici MCP kontrolü etkinleştirilsin mi?',
        enableConfirmDesc:
          'MCP etkinleştirildiğinde PicoClaw durdurulur ve tüm etkin PicoClaw oturumları kapatılır.',
        failed: 'MCP işlemi başarısız oldu',
        copyFailed: 'Kopyalama başarısız. Elle kopyalayın.',
        okBtn: 'Onayla',
        cancelBtn: 'İptal',
        showKey: 'Anahtarı göster',
        hideKey: 'Anahtarı gizle',
        regenerateKey: 'Anahtarı yeniden oluştur'
      },
      redfish: {
        example: 'Örnek',
        title: 'Redfish',
        service: 'Redfish hizmeti',
        serviceDesc:
          "redfishtool ve Ansible gibi araçlardan güç kontrolü, sanal medya ve durum bilgisi için DMTF Redfish API'si. Kapatmak tüm Redfish oturumlarını sonlandırır.",
        endpoint: 'Hizmet kökü',
        httpsOn: 'Kart, çoğu Redfish aracının ihtiyaç duyduğu HTTPS ile hizmet veriyor.',
        httpsOff:
          'Kart şifrelenmemiş HTTP ile hizmet veriyor. Çoğu Redfish aracı HTTPS gerektirir: "Ayarlar > Ağ" bölümünden açın.',
        credentials:
          'Redfish, KVM hesaplarını Basic kimlik doğrulaması veya Redfish oturumuyla, ayrıca X-Auth-Token olarak gönderilen API anahtarlarını kabul eder. API anahtarları API Anahtarları sayfasından yönetilir.',
        powerActions: 'Güç işlemleri',
        powerActionsDesc:
          'Şu anda sunulan sıfırlama türleri. On, ForceOff ve GracefulShutdown güç durumunu bilmeyi gerektirir; bu yüzden yalnızca güç menüsünde "Güç LED\'i bağlı" açıkken sunulur.',
        sessions: 'Oturumlar',
        noSessions: 'Açık Redfish oturumu yok',
        created: 'Oluşturulma',
        lastUsed: 'Son kullanım',
        refresh: 'Yenile',
        end: 'Sonlandır',
        endConfirmTitle: 'Bu Redfish oturumu sonlandırılsın mı?',
        endConfirmDesc:
          "Oturumun token'ı hemen geçersiz olur. İstemcinin yeniden giriş yapması gerekir.",
        failed: 'Redfish işlemi başarısız oldu',
        copyFailed: 'Kopyalama başarısız oldu. Elle kopyalayın.',
        okBtn: 'Onayla',
        cancelBtn: 'İptal'
      },
      ipmi: {
        copyBeforeSave: 'Parolayı şimdi kopyalayın. Kaydedildikten sonra tekrar gösterilemez.',
        noLogin:
          'IPMI açık, ancak hiçbir etkin hesabın IPMI parolası yok, bu yüzden kimse oturum açamaz. Aşağıdan bir tane belirleyin.',
        title: 'IPMI',
        warning:
          "IPMI kimlik doğrulaması tasarımı gereği zayıftır. Karta erişebilen ve bir kullanıcı adını bilen herkes, o kullanıcının IPMI parolasının karmasını alıp çevrimdışı kırmayı deneyebilir. Üretilmiş parolalar kullanın, IPMI'yi yalnızca güvenilir bir ağda açın ve araç destekliyorsa HTTPS üzerinden Redfish'i tercih edin.",
        service: 'LAN üzerinden IPMI',
        serviceDesc:
          'Ana makinenin gücü ve durumu için UDP 623 numaralı bağlantı noktasında IPMI 2.0 (RMCP+, ipmitool lanplus). IPMI 1.5 ve 0 numaralı şifre takımı reddedilir. Kapatmak tüm IPMI oturumlarını sonlandırır.',
        example: 'Örnek',
        copyFailed: 'Kopyalama başarısız. Elle kopyalayın.',
        ledOn: 'Güç durumu, on, off, soft, cycle ve reset kullanılabilir.',
        ledOff:
          'Güç menüsünde "Güç LED\'i bağlı" kapalı, bu yüzden güç durumu bilinmiyor. Yalnızca "power reset" çalışır: status, on, off, soft ve cycle reddedilir.',
        accounts: 'Hesaplar',
        accountsDesc:
          'IPMI, KVM hesaplarıyla oturum açar; her hesabın web parolasından ayrı kendi IPMI parolası vardır. Yöneticiler ADMINISTRATOR alır. Kullanıcılar USER alır: "-L USER" ile güç durumunu okuyabilir ama değiştiremezler.',
        passwordSet: 'IPMI parolası ayarlı',
        passwordNotSet: 'IPMI parolası yok: IPMI ile oturum açamaz',
        nameTooLong: 'Ad 16 karakterden uzun, IPMI buna izin vermez',
        accountDisabled: 'Hesap devre dışı',
        setPassword: 'Parola ayarla',
        changePassword: 'Parolayı değiştir',
        remove: 'Kaldır',
        removeConfirmTitle: '{{user}} için IPMI parolası kaldırılsın mı?',
        removeConfirmDesc: 'Hesap artık IPMI ile oturum açamaz ve IPMI oturumları sona erer.',
        passwordTitle: '{{user}} için IPMI parolası',
        passwordDesc:
          'Web parolasından farklı, 12 ile 20 arası yazdırılabilir ASCII karakter. IPMI, kartın parolayı geri okuyabileceği bir biçimde saklamasını gerektirir, bu yüzden başka hiçbir yerde kullanılmayan bir parola seçin. Kaydetmeden önce kopyalayın: bir daha gösterilmez.',
        passwordPlaceholder: 'IPMI parolası',
        generate: 'Üret',
        copy: 'Kopyala',
        save: 'Kaydet',
        passwordLength: '12 ile 20 arası karakter kullanın.',
        passwordChars: 'Yalnızca yazdırılabilir ASCII karakterler kullanın.',
        saved: 'IPMI parolası kaydedildi',
        failed: 'IPMI işlemi başarısız oldu',
        okBtn: 'Onayla',
        cancelBtn: 'İptal'
      },
      ssh: {
        service: 'SSH sunucusu',
        serviceDesc: "sshd'yi şimdi ve her açılışta başlat",
        failed: 'SSH ayarları yüklenemedi',
        rootDefault: 'root hâlâ fabrika parolasını kullanıyor',
        rootEmpty: "root'un parolası yok",
        rootWarning:
          "Konsola veya SSH'ye erişen herkes root olarak girebilir. {{account}} > {{password}} altında bir parola belirleyin: cihaz sahibi için bu, root parolasını da belirler.",
        connection: 'Bağlantı',
        command: 'root olarak giriş yap',
        port: 'Port',
        viaVpn: '{{name}} üzerinden',
        notRunning: 'sshd çalışmıyor. Bağlanmak için SSH sunucusunu açın.',
        hostKeys: 'Ana makine anahtarı parmak izleri',
        hostKeysDesc: 'Bunları ssh ilk bağlantıda gösterdiğiyle karşılaştırın.',
        noHostKeys: 'Henüz ana makine anahtarı yok. sshd bunları ilk açılışta oluşturur.',
        keys: 'Yetkili anahtarlar',
        keysDesc:
          'root olarak girebilen açık anahtarlar. Veri bölümünde saklanırlar, bu yüzden güncellemeler onları korur.',
        noKeys: 'Henüz yetkili anahtar yok.',
        noComment: 'yorum yok',
        addPlaceholder: 'Bir açık anahtar yapıştırın, örneğin ~/.ssh/id_ed25519.pub içeriği',
        add: 'Anahtar ekle',
        added: 'Anahtar eklendi',
        removed: 'Anahtar kaldırıldı',
        deleteConfirm: 'Bu anahtar kaldırılsın mı?',
        deleteConfirmDesc: 'Artık giriş yapamaz. Açık oturumlar açık kalır.',
        invalidKey: 'Bu bir açık anahtar değil. Bir .pub dosyasından tek bir satır yapıştırın.',
        keyOptions: 'command= veya from= gibi seçenekler içeren anahtarlar burada kabul edilmez.',
        duplicateKey: 'Bu anahtar zaten yetkili.',
        lastKey: 'Yalnızca anahtarla giriş açıkken son anahtar kaldırılamaz.',
        keysOnly: 'Yalnızca anahtarlar',
        keysOnlyDesc:
          'Parola ve keyboard-interactive ile girişi kapatın. Açık oturumlar açık kalır.',
        keysOnlyNeedsKey: 'Önce yetkili bir anahtar ekleyin, yoksa kimse giriş yapamaz.',
        keysOnlyOn: 'Parolayla giriş kapatıldı',
        keysOnlyOff: 'Parolayla giriş açıldı',
        notHonoured: 'Bu imajdaki sshd bu ayarı okumuyor, bu yüzden parolayla giriş açık kalıyor.',
        reloadFailed:
          'Kaydedildi, ancak sshd yeniden yüklenemedi. sshd bir sonraki başlatılışında uygulanır.',
        notApplied:
          'sshd hâlâ parola kabul ediyor. Ayarı uygulamak için SSH sunucusunu kapatıp açın.',
        changePort: 'Değiştir',
        portConfirm: 'SSH bağlantı noktası {{port}} olarak değiştirilsin mi?',
        portConfirmDesc:
          'Mevcut SSH oturumlarınız açık kalır. Yeni bağlantılar {{port}} bağlantı noktasını kullanmalıdır. Güvenlik duvarınızın buna izin verdiğinden emin olun.',
        portChanged: 'SSH bağlantı noktası {{port}} olarak değiştirildi',
        portInvalid: '1 ile 65535 arasında bir bağlantı noktası girin.',
        portReserved: "Bu bağlantı noktasını IronKVM'in kendisi kullanıyor. Başka birini seçin.",
        portInUse: 'IronKVM üzerindeki başka bir program bu bağlantı noktasını zaten dinliyor.',
        portNotHonoured: 'Bu imajdaki sshd bu ayarı okumuyor, bu yüzden bağlantı noktası değişmedi.'
      },
      vnc: {
        address: 'Adres',
        certHint:
          "VeNCrypt X509Plain cihazın kendinden imzalı sertifikasını kullanır, bu yüzden istemci ilk bağlantıda uyarır. Kabul edin ya da sertifikayı bu sayfanın HTTPS adresinden kaydedip TigerVNC'ye -X509CA=<dosya> ile verin.",
        title: 'VNC',
        service: 'VNC sunucusu',
        serviceDesc:
          'TigerVNC veya Remmina gibi bir VNC istemcisinin ana bilgisayarı görmesini ve denetlemesini sağlar. İstemci Tight kodlamasını desteklemelidir. Aynı anda tek oturum.',
        credentials:
          'Bir KVM hesabıyla oturum açın. Bağlantı, kartın TLS sertifikasıyla şifrelenir (VeNCrypt X509Plain).',
        port: 'Bağlantı noktası',
        portDesc: 'Sunucunun dinlediği TCP bağlantı noktası.',
        maxFps: 'Kare hızı sınırı',
        maxFpsDesc: 'Bir istemciye gönderilen en fazla saniyedeki kare sayısı.',
        vncAuth: 'Basit VNC kimlik doğrulaması',
        vncAuthDesc:
          'VeNCrypt desteği olmayan istemciler için. Hesap yerine ayrı bir VNC parolasını denetler.',
        vncAuthWarning:
          'Basit VNC kimlik doğrulaması bağlantıyı şifrelemez. Ağ yolundaki herkes ekranı ve tuş vuruşlarını görebilir. Yalnızca güvenilir bir ağda kullanın.',
        password: 'VNC parolası',
        passwordSet: 'Bir parola ayarlı. Değiştirmek için yenisini yazın.',
        passwordInvalid: 'VNC parolası 6 ile 8 karakter arasında olmalıdır.',
        save: 'Kaydet',
        saved: 'Ayarlar kaydedildi',
        state: 'Durum',
        listening: '{{port}} numaralı bağlantı noktasında dinliyor',
        notListening: 'Dinlemiyor',
        noSession: 'Açık oturum yok',
        client: 'İstemci',
        user: 'Kullanıcı',
        method: 'Kimlik doğrulama',
        methodVencrypt: 'TLS üzerinden hesap',
        methodVnc: 'VNC parolası',
        since: 'Bağlantı zamanı',
        resolution: 'Çözünürlük',
        framesSent: 'Gönderilen kareler',
        lastError: 'Son oturum sona erdi: {{error}}',
        refresh: 'Yenile',
        disconnect: 'Bağlantıyı kes',
        disconnectConfirmTitle: 'VNC oturumu sonlandırılsın mı?',
        disconnectConfirmDesc:
          'İstemcinin bağlantısı hemen kesilir ve basılı tuttuğu tüm tuşlar ve düğmeler bırakılır.',
        failed: 'VNC işlemi başarısız oldu',
        okBtn: 'Onayla',
        cancelBtn: 'İptal'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Ana makine watchdog',
        serviceDesc:
          'Ana makinenin açık olması gerekirken görüntüsü zaman aşımı boyunca değişmezse veya HDMI sinyali yoksa, kart reset düğmesine basar ya da ana makineyi kapatıp yeniden açar.',
        stillWarning:
          'Ekranı uyku moduna geçen veya çalışırken görüntüsü sabit kalan bir ana makine donmuş görünür. Ana makinede ekran uykusunu kapatın veya bir ping adresi girin.',
        ledHint:
          'Güç menüsünde "Güç LED\'i bağlı" kapalı. Watchdog ana makinenin ne zaman kapalı olduğunu göremez, bu yüzden onu her zaman açık sayar.',
        timeout: 'Zaman aşımı',
        timeoutDesc:
          'Watchdog devreye girmeden önce ana makinenin ne kadar süre yaşam belirtisi göstermeyebileceği.',
        action: 'Eylem',
        actionDesc: 'Güç döngüsü güç düğmesini 5 saniye basılı tutar, ardından yeniden basar.',
        actionReset: 'Sıfırla',
        actionPower: 'Güç döngüsü',
        cooldown: 'Bekleme süresi',
        cooldownDesc: 'İki eylem arasındaki en kısa süre.',
        maxPerHour: 'Saat başına eylem',
        maxPerHourDesc: 'Bir saatteki en fazla eylem sayısı.',
        pingHost: 'Ping adresi',
        pingHostDesc:
          'Ana makinenin IP adresi. Bir yanıt yaşam belirtisi sayılır. Ping atmamak için boş bırakın.',
        pingHostInvalid: 'Bir IPv4 veya IPv6 adresi girin.',
        minutes: 'dk',
        save: 'Kaydet',
        saved: 'Kaydedildi',
        state: 'Algılayıcı',
        status: {
          off: 'Kapalı',
          watching: 'İzliyor',
          hostOff: 'Ana makine kapalı',
          captureOff: 'HDMI yakalama kapalı',
          cooldown: 'Beklemede',
          capped: 'Saatlik sınıra ulaşıldı',
          acting: 'Devrede'
        },
        signal: 'HDMI sinyali',
        yes: 'Evet',
        no: 'Hayır',
        led: 'Güç LED’i',
        on: 'Yanıyor',
        off: 'Sönük',
        ledNotConnected: 'Bağlı değil',
        ping: 'Ping',
        pingNotSet: 'Ayarlanmadı',
        pingReply: 'Yanıt veriyor',
        pingNoReply: 'Yanıt yok',
        lastChange: 'Son görüntü değişikliği',
        never: 'Hiç',
        actsIn: 'Devreye girmesine kalan',
        actionsLastHour: 'Son bir saatteki eylemler',
        duration: '{{minutes}} dk {{seconds}} sn',
        log: 'Günlük',
        noLog: 'Watchdog henüz devreye girmedi.',
        refresh: 'Yenile',
        reasonFrozen: 'Görüntü değişmedi',
        reasonNoSignal: 'HDMI sinyali yok',
        stuckFor: '{{duration}} boyunca yaşam belirtisi yok',
        pressFailed: 'Düğmeye basılamadı: {{error}}',
        noScreenshot: 'Ekran görüntüsü yok',
        failed: 'Watchdog işlemi başarısız oldu',
        powerNeedsLed: 'Güç döngüsü için güç menüsünde "Güç LED\'i bağlı" gerekir.',
        noLedConfirmTitle: "Watchdog güç LED'i olmadan açılsın mı?",
        noLedConfirmDesc:
          "Kart, ana makinenin ne zaman kapalı olduğunu göremez, bu yüzden onu hep açık sayar. Ana makineyi kapatırsanız, süre dolunca watchdog sıfırlamaya basar. Bunu önlemek için güç LED'ini bağlayın.",
        noLedConfirmOk: 'Aç',
        cancel: 'İptal'
      },
      media: {
        title: 'Sanal medya',
        description:
          'Araç çubuğundaki Medya penceresinin kurulumu. Kalıpları bağlama, ekleme ve Ventoy kümesini seçme pencerede yapılır.',
        ejectFirst: 'Ventoy diski bir sürücüde. Kaldırmak için Medya penceresinden çıkarın.'
      },
      netboot: {
        title: 'Ağdan önyükleme',
        isoDownload: 'İndir',
        description:
          "Ana makineyi ağdan başlatın: USB ağ bağlantısı üzerinden iPXE ve KVM'deki görüntülerin menüsü ya da LAN'da proxy DHCP ile netboot.xyz.",
        addon: 'dnsmasq ve önyükleme dosyaları',
        addonDesc:
          "/data'ya kurulur: dnsmasq Alpine'den, iPXE ve netboot.xyz kendi sürümlerinden gelir, her biri sağlama toplamıyla doğrulanır.",
        install: 'Kur',
        installing: 'Kuruluyor. Bu birkaç dakika sürebilir.',
        uninstall: 'Kaldır',
        uninstallConfirm:
          'Ağdan önyükleme kapatılsın ve dnsmasq ile önyükleme dosyaları kaldırılsın mı?',
        needsData: 'Ağdan önyükleme, /data bölümü bağlı bir IronKVM görüntüsü gerektirir.',
        usb: 'USB ağ bağlantısında',
        usbDesc:
          "USB ağ bağlantısı açıkken ona udhcpd yerine dnsmasq hizmet verir. Ana makine tek adresini yönlendirici ve DNS sunucusu olmadan, mimarisine uygun iPXE'yi ve KVM'deki ISO görüntülerinin menüsünü alır.",
        linkOff: 'USB ağ bağlantısı kapalı. Aygıt, USB ağı altından açın.',
        menuUrl: 'Menü',
        leases: 'Ana makinenin kirası',
        noLeases: 'Henüz yok',
        netbootxyzNote:
          'Menüdeki netboot.xyz internetten yüklenir ve USB bağlantısı internete ulaşmaz. Ana makinenin bunun için başka bir ağ bağlantı noktasında internete ihtiyacı vardır.',
        lan: "LAN'da proxy DHCP",
        lanDesc:
          "LAN'daki PXE istemcilerine netboot.xyz sunar; netboot.xyz de menüsünü internetten yükler. Asla adres dağıtmaz ve KVM'deki görüntüleri sunmaz.",
        lanWarning:
          "netboot.xyz yalnızca ana makineye değil, bu LAN'daki her PXE istemcisine sunulur. Bunu yalnızca kontrol ettiğiniz bir ağda açın.",
        lanConfirm: "LAN'da proxy DHCP açılsın mı?",
        lanInterface: 'LAN',
        running: 'Çalışıyor',
        stopped: 'Çalışmıyor',
        images: 'Menüdeki görüntüler',
        noImages: 'Görüntü dizininde ISO görüntüsü yok.',
        boots: 'Son önyüklemeler',
        noBoots: 'Ana makine henüz hiçbir şey almadı.',
        log: 'dnsmasq günlüğü',
        refresh: 'Yenile',
        okBtn: 'Onayla',
        cancelBtn: 'İptal',
        failed: 'Ağdan önyükleme işlemi başarısız oldu'
      },
      about: {
        title: 'IronKVM Hakkında',
        information: 'Bilgi',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Uygulama sürümü',
        applicationTip: 'IronKVM web uygulaması sürümü',
        image: 'İmaj Sürümü',
        imageTip: 'IronKVM kart görüntüsü ve üzerine kurulduğu NanoKVM sistem görüntüsü',
        kernel: 'Çekirdek Sürümü',
        kernelTip: 'Şu anda çalışan Linux çekirdeğinin sürümü',
        deviceKey: 'Cihaz Anahtarı',
        community: 'Topluluk',
        hostname: 'Ana makine adı',
        hostnameUpdated: 'Hostname güncellendi. Uygulamak için yeniden başlatın.',
        ipType: {
          Wired: 'Kablolu bağlantı',
          Wireless: 'Kablosuz bağlantı',
          Other: 'Diğer'
        },
        hostnameInvalid:
          'Harf, rakam ve kısa çizgi kullanın; noktayla ayrılan her bölümde en fazla 63. Bir bölümün başında veya sonunda kısa çizgi olamaz.',
        hostnameFailed: 'Ana bilgisayar adı değiştirilemedi',
        editHostname: 'Ana bilgisayar adını düzenle',
        docs: 'Belgeler',
        hardware: 'Donanım',
        hardwareFaq: 'Donanım SSS',
        disclaimer:
          'IronKVM: Sipeed NanoKVM için sağlamlaştırılmış topluluk yazılımı. Sipeed ile bağlantılı değildir.',
        basedOn: 'NanoKVM {{version}} tabanlı'
      },
      preferences: {
        title: 'Tercihler'
      },
      performance: {
        title: 'Performans',
        memory: {
          title: 'Bellek',
          description: 'RAM, takas alanı ve bunları kullananlar. Birkaç saniyede bir güncellenir.',
          ram: 'RAM',
          of: '{{used}} / {{total}}',
          available: '{{available}} kullanılabilir',
          availableLow:
            'Yalnızca {{available}} kullanılabilir. Hizmetler yavaşlayabilir veya durdurulabilir.',
          swap: 'Takas alanı',
          swapFile: 'Takas dosyası',
          zram: 'Sıkıştırılmış takas',
          zramRam: "RAM'de {{ram}}",
          consumers: 'Başlıca tüketiciler',
          addons: 'Eklentiler',
          addonsTip: "Tailscale ve NetBird'ün çalıştığı bellek grubu, sınırına göre.",
          video: 'Video Belleği',
          videoTip: 'Video yakalama için ayrılmış bellek. Sistemin geri kalanıyla paylaşılmaz.',
          videoGenerations_one: '{{count}} önceki IronKVM oturumu video belleğini tutuyor',
          videoGenerations_other: '{{count}} önceki IronKVM oturumu video belleğini tutuyor',
          videoReboot: 'Geri kazanmak için yeniden başlatın.'
        }
      },
      appearance: {
        thisBrowser: 'Bu tarayıcı',
        thisBrowserDesc:
          'Yalnızca bu tarayıcıda saklanır. Diğer tarayıcılar kendi ayarlarını kullanır.',
        deviceWide: 'Cihaz',
        deviceWideDesc: 'Cihazda saklanır. Cihazı açan herkes için geçerlidir.',
        language: 'Dil',
        languageDesc: 'Arayüz için dili seçin',
        webTitle: 'Site başlığı',
        webTitleDesc: 'Görünen site başlığını güncelleyin',
        menuBar: {
          title: 'Menü Çubuğu',
          mode: 'Görüntüleme Modu',
          modeDesc: 'Ekranda menü çubuğunu görüntüle',
          modeOff: 'Kapalı',
          modeAuto: 'Otomatik gizle',
          modeAlways: 'Her zaman görünür',
          keyboardLedStatus: 'Klavye kilidi göstergeleri',
          keyboardLedStatusDesc:
            'Uzak bilgisayarın Num Lock, Caps Lock ve Scroll Lock durumunu göster',
          icons: 'Alt Menü Simgeleri',
          iconsDesc: 'Menü çubuğunda alt menü simgelerini görüntüle'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Uzak klavye kilidi durumu',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Açık',
        off: 'Kapalı',
        unknown: 'Bilinmiyor'
      },
      device: {
        title: 'Cihaz',
        oled: {
          title: 'OLED',
          description: 'Oled ekranı ... sonra kapatın',
          brightness: 'OLED parlaklığı',
          brightnessDescription: 'Daha düşük seviye ekranın ömrünü uzatır',
          brightnessLevels: {
            '64': 'En düşük',
            '96': 'Düşük',
            '128': 'Orta',
            '160': 'Yüksek',
            '207': 'Varsayılan',
            '255': 'En yüksek'
          },
          0: 'Hiçbir zaman',
          15: '15 saniye',
          30: '30 saniye',
          60: '1 dakika',
          180: '3 dakika',
          300: '5 dakika',
          600: '10 dakika',
          1800: '30 dakika',
          3600: '1 saat'
        },
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Ön panel',
          roomMic: 'Oda mikrofonu'
        },
        roomMic: {
          allow: 'Oda mikrofonuna izin ver',
          description:
            "İzleyiciler odayı KVM'nin yerleşik mikrofonuyla dinleyebilir. Açıkken izleyen herkes görür.",
          unavailable: 'Bu çekirdekte kullanılamaz',
          gain: 'Mikrofon kazancı'
        },
        cpuFreq: {
          title: 'CPU Frekansı',
          description: 'Bir sonraki açılışta uygulanacak CPU saat hızını ayarlayın',
          tip: 'CPU 850 MHz ile açılır ve 1000 MHz için derecelendirilmiştir. Yeni değer sistem çalışırken değil, bir sonraki açılışta uygulanır. 1000 MHz spesifikasyon dahilindedir; sıcaklık her iki ayarda da sınırların oldukça altında kalır.',
          running: 'Çalışan: {{mhz}} MHz',
          rebootToApply: 'uygulamak için yeniden başlatın',
          rebootConfirm: '{{mhz}} MHz uygulamak için şimdi yeniden başlatılsın mı?'
        },
        swap: {
          title: 'Swap',
          disable: 'Aktifleştir',
          description: 'Swap dosyasının boyutunu belirle',
          tip: 'Bu özelliği aktifleştirmek micro SD kartınızın ömrünü kısaltabilir!',
          active: 'Aktif - {{used}} / {{total}}',
          inactive: 'Ayarlandı, ancak kullanılmıyor'
        },
        zram: {
          title: 'Sıkıştırılmış swap (zram)',
          description: "Swap, SD kart yerine sıkıştırılmış RAM'de",
          tip: "zram, swap'ı SD karttan uzak tutar, bu yüzden karta aşınma yapmaz. Arkasında disk swap'ı yoktur: zram dolarsa çekirdek yavaşça sayfalamak yerine bir işlemi durdurur. Bellek sınırı, zram'ın ne kadar RAM kullanabileceğini belirler.",
          unavailable: 'Bu cihazda çekirdek modülleri yüklü değil',
          inactive: 'Etkin, ancak aygıt başlamadı',
          active: 'Aktif - {{used}} / {{total}}, {{ratio}}x',
          off: 'Kapalı',
          detail: {
            algorithm: 'Algoritma: {{algorithm}}',
            memory: 'Kullanılan bellek: {{used}} / {{limit}}',
            memoryNoLimit: 'Kullanılan bellek: {{used}}, sınır ayarlanmadı',
            counters:
              "Swap'tan okunan sayfalar {{in}}, swap'a yazılan sayfalar {{out}} (tüm swap aygıtları, açılıştan beri)"
          }
        },
        mouseJiggler: {
          title: 'Fare Oynatıcı',
          description: 'Uzak ana bilgisayarın uykuya geçmesini engeller',
          disable: 'Devre dışı bırak',
          absolute: 'Mutlak mod',
          relative: 'Bağıl mod'
        },
        mdns: {
          description: 'mDNS keşif hizmetini etkinleştir',
          tip: 'Kullanmıyorsanız devre dışı bırakabilirsiniz'
        },
        hdmi: {
          description: 'HDMI/Momitör çıktısını aktifleştir',
          idleTimeoutTitle: 'Etkin olmayan yakalama zaman aşımı',
          idleTimeoutDescription:
            'Etkin görüntüleyici olmadığında HDMI yakalamayı şu süre sonunda durdur:',
          minutes: 'dk'
        },
        hidOnly: 'Yalnızca HID modu',
        hidOnlyDesc:
          'Yalnızca temel HID kontrolünü koruyarak sanal aygıtları taklit etmeyi bırakın',
        disk: 'Sanal Disk',
        diskDesc: "Sanal U-disk'i uzak ana bilgisayara bağla",
        network: 'Sanal Ağ',
        networkDesc: 'Sanal ağ kartını uzak ana bilgisayara bağla',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Ana makine:',
          description:
            'USB kablosu üzerinden uzak ana bilgisayarla özel bir ağ bağlantısı. Ana bilgisayar ağ geçidi ve DNS olmadan bir adres alır, bu yüzden IronKVM üzerinden yerel ağınıza ulaşamaz.',
          mode: 'Protokol',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (NCM desteği olmayan ana bilgisayarlar için)',
          rndis: 'RNDIS (artık sunulmuyor)',
          rndisNote: 'Bu bağlantı artık sunulmayan RNDIS kullanıyor. NCM veya ECM seçin.',
          subnet: 'Alt ağ',
          subnetDesc:
            '/24 ile /30 arasında özel bir IPv4 ağı. IronKVM ilk adresi, ana bilgisayar ikinci adresi alır.',
          invalidSubnet: '172.31.255.0/30 gibi bir alt ağ girin.',
          apply: 'Uygula',
          confirm: 'USB aygıtı yeniden bağlansın mı?',
          reenumerate:
            'Uygulamak USB bağlantısını yeniden kurar. Ana bilgisayar birkaç saniye boyunca klavyeyi, fareyi ve sanal diski kaybeder.'
        },
        audio: 'Sanal Hoparlör',
        audioDesc:
          'Uzak ana bilgisayara bir USB ses kartı sunar, böylece sesini duyabilirsiniz. Ana bilgisayarın bunu çıkış aygıtı olarak seçmesi gerekir. Bunu değiştirmek USB bağlantısını yeniden kurar.',
        audioNote:
          "Ses her iki H.264 modunda (WebRTC ve Direct) kullanılabilir, MJPEG'de kullanılamaz",
        console: 'Seri Konsol',
        consoleDesc:
          "Ağa erişilemediğinde bu IronKVM'e giriş yapabilmek için uzak ana bilgisayara bir USB seri port sunar",
        consoleTip:
          'Uzak ana bilgisayarı kontrol eden herkes bu IronKVM için bir giriş istemi görür. Etkinleştirmeden önce güçlü bir şifre belirleyin (Hesap - Şifremi Değiştir).',
        usbApply: {
          changed: 'Değiştirildi',
          discard: 'Vazgeç',
          pending: 'Değişiklikler henüz uygulanmadı.'
        },
        endpoints: {
          title: 'USB yuvaları',
          free: '{{total}} yuvadan {{free}} boş',
          slots: 'Yuva: {{count}}',
          full: 'Yeterli boş USB yuvası yok. Önce başka bir şeyi kapatın.',
          inactive:
            'Açık ama çalışmıyor: USB denetleyicisinin yuvaları bitti. Başka bir aygıtı kapatın, bu hemen başlar.',
          explain:
            'USB denetleyicisinin sabit sayıda yuvası (giriş uç noktası) vardır ve klavye ile fare her zaman bir kısmını kullanır. Sığabilecek olandan fazla aygıt açıksa klavye ve fare kalır, geri kalanı kapatılır.',
          error: 'Cihaza ulaşılamadı. Tekrar deneyin.',
          fitTogether: 'Birlikte sığanlar: {{sets}}'
        },
        reboot: 'Yeniden Başlat',
        rebootDesc: "IronKVM'i yeniden başlatmak istediğinizden emin misiniz?",
        okBtn: 'Evet',
        cancelBtn: 'Hayır',
        rebootFailed: 'Yeniden başlatma başarısız'
      },
      network: {
        title: 'Ağ',
        wifi: {
          disconnectBtn: 'Bağlantıyı kes',
          disconnectWarning:
            "IronKVM'e bu Wi-Fi ağı üzerinden erişiyorsanız bu sayfanın bağlantısı kesilir.",
          disconnected: 'Wi-Fi bağlantısı kesildi',
          title: 'Wi-Fi',
          description: 'Wi-Fi ayarlayın',
          apMode: "AP modu etkin, QR kodu tarayarak Wi-Fi'ye bağlanın",
          connect: "Wi-Fi'ye bağlan",
          connectDesc1: "Lütfen ağ SSID'sini ve parolayı girin",
          connectDesc2: 'Bu ağa katılmak için parolayı girin',
          disconnect: 'Ağ bağlantısını kesmek istediğinizden emin misiniz?',
          failed: 'Bağlantı başarısız, lütfen tekrar deneyin.',
          ssid: 'Ad',
          password: 'Parola',
          joinBtn: 'Katıl',
          confirmBtn: 'Tamam',
          cancelBtn: 'İptal'
        },
        tls: {
          description: 'HTTPS protokolünü etkinleştir',
          tip: 'HTTPS protokolü bağlantıda gecikmeye sebep olabilir, özellikle MJPEG görüntü modu ile.',
          restarting: 'Cihaz sunucusu yeniden başlatılıyor, bu yaklaşık iki dakika sürer...',
          waiting: 'Cihazın yeniden yanıt vermesi bekleniyor...',
          waitingHttp: "http'ye geri dönülüyor. Sayfa kendiliğinden açılmazsa yeniden yükleyin.",
          failed: 'HTTPS ayarı değiştirilemedi',
          enableConfirm: 'HTTPS açılsın mı?',
          disableConfirm: 'HTTPS kapatılsın mı?',
          confirmDesc:
            'Bu, oturumunuzu kapatır ve cihaz sunucusunu yeniden başlatır; yaklaşık iki dakika sürer. Ardından sayfa {{url}} adresini açar.',
          confirmOk: 'Devam',
          confirmCancel: 'İptal'
        },
        ethernet: {
          title: 'IP Adresi',
          description: 'IronKVM cihazının kablolu ağdaki adresini nasıl aldığını yapılandırın',
          dhcp: 'DHCP',
          manual: 'Manuel',
          networkDetails: 'Ağ Ayrıntıları',
          interface: 'Arayüz',
          ipAddress: 'IP Adresi',
          subnetMask: 'Alt Ağ Maskesi',
          router: 'Yönlendirici',
          save: 'Uygula',
          invalidAddress: 'Geçerli bir IP adresi girin',
          invalidMask: 'Geçerli bir alt ağ maskesi girin, örneğin 255.255.255.0 veya 24',
          invalidRouter: 'Geçerli bir yönlendirici adresi girin',
          addressRequired: 'IP adresi gereklidir',
          maskRequired: 'Alt ağ maskesi gereklidir',
          applyTitle: 'IronKVM adresi değiştirilsin mi?',
          applyWarning:
            'Bu sayfayla bağlantı kesilecek. IronKVM yeni adresi uygular ve ona bu adresten ulaşmanız için {{seconds}} saniye bekler. Ona ulaşmak değişikliği korur. Ona hiçbir şey ulaşmazsa IronKVM önceki ayarları geri yükler.',
          applyConfirm: 'Uygula',
          applyCancel: 'İptal',
          applyFailed: 'Adres uygulanamadı',
          trialTitle: 'Onay bekleniyor',
          trialDhcp: 'IronKVM, DHCP üzerinden adres istiyor.',
          trialStatic: 'IronKVM şimdi {{address}} adresinde.',
          trialInstruction:
            "IronKVM'i yeni adresinde açın ve isterse oturum açın. Ona orada ulaşmak değişikliği korur. {{seconds}} saniye içinde IronKVM'e hiçbir şey ulaşmazsa önceki ayarlar geri yüklenir.",
          trialOpen: 'Yeni adresi aç',
          trialKeep: 'Bu ayarları koru',
          trialKept: 'Yeni adres kaydedildi',
          trialKeepFailed: 'Ayarlar korunamadı',
          trialGone: 'Değişiklik zaten geri alındı. Tekrar deneyin.',
          unsaved: 'Kaydedilmemiş değişiklikler'
        },
        dns: {
          title: 'DNS',
          description: 'IronKVM için DNS sunucularını yapılandır',
          mode: 'Mod',
          dhcp: 'DHCP',
          manual: 'Manuel',
          add: 'DNS ekle',
          save: 'Kaydet',
          invalid: 'Geçerli bir IP adresi girin',
          noDhcp: 'Şu anda DHCP DNS mevcut değil',
          saved: 'DNS ayarları kaydedildi',
          saveFailed: 'DNS ayarları kaydedilemedi',
          unsaved: 'Kaydedilmemiş değişiklikler',
          maxServers: 'En fazla {{count}} DNS sunucusuna izin verilir',
          dnsServers: 'DNS Sunucuları',
          dhcpServersDescription: "DNS sunucuları DHCP'den otomatik olarak alınır",
          manualServersDescription: 'DNS sunucuları manuel olarak düzenlenebilir',
          networkDetails: 'Ağ Ayrıntıları',
          interface: 'Arayüz',
          ipAddress: 'IP Adresi',
          subnetMask: 'Alt Ağ Maskesi',
          router: 'Yönlendirici',
          none: 'Yok'
        },
        syslog: {
          title: 'Uzak günlükleme',
          description:
            'Sistem, çekirdek ve IronKVM günlüklerini UDP ile bir syslog toplayıcısına gönderir.',
          placeholder: 'ana makine veya ana makine:port, UDP, varsayılan port 514',
          save: 'Kaydet',
          turnOff: 'Kapat',
          test: 'Test mesajı gönder',
          sent: '<v>{{message}}</v> gönderildi. Toplayıcıda bunu arayın.',
          forwarding: '<v>{{target}}</v> adresine iletiliyor',
          local: 'Yalnızca yerel',
          savedButLocal:
            '<v>{{target}}</v> kaydedildi, ancak çalışan günlükleyici yalnızca yerel kaydediyor',
          savedButForwarding:
            '<v>{{target}}</v> kaydedildi, ancak çalışan günlükleyici <v>{{active}}</v> adresine iletiyor',
          offButForwarding:
            'Kapatıldı, ancak çalışan günlükleyici hâlâ <v>{{active}}</v> adresine iletiyor',
          unsupported: 'Bu imajın günlükleyicisi ayarı yok sayıyor; kullanmak için güncelleyin',
          loadFailed: 'Günlük ayarı okunamadı',
          saveFailed: 'Günlük ayarı kaydedilemedi',
          testFailed: 'Test mesajı gönderilemedi',
          metrics: 'Metrikler',
          metricsDesc:
            "Prometheus aşağıdaki URL'yi <link>API Anahtarları</link> içinden bir bearer token ile okuyabilir.",
          metricsUrl: "Metrik URL'si",
          errors: {
            empty: 'Bir ana makine veya ana makine:port girin',
            long: 'Ana makine adı için çok uzun',
            ipv6: 'IPv6 adresini köşeli parantez içine alın, örneğin [fd00::1]:514',
            brackets: 'Parantez içindeki adres bir IPv6 adresi değil',
            host: 'Yalnızca harf, rakam, nokta ve tire kullanın; nokta veya tireyle başlamayın',
            port: 'Port 1 ile 65535 arasında bir sayı olmalı'
          }
        }
      },
      vpn: {
        kvmUrl: 'KVM adresi',
        moreTip: 'Diğer işlemler',
        updateTip: '{{version}} sürümüne güncelle',
        loading: 'Yükleniyor...',
        okBtn: 'Evet',
        cancelBtn: 'Hayır',
        restart: '{{name}} yeniden başlatılsın mı?',
        update: '{{name}}, {{version}} sürümüne güncellensin mi?',
        updateDesc: 'Arka plan hizmeti çalışıyorsa yeniden başlar. Oturum bilgisi korunur.',
        notInstall: '{{name}} yüklü değil.',
        install: 'Yükle',
        installing: 'Yükleniyor',
        installFailed: 'Yükleme başarısız oldu',
        retry: 'Tekrar dene',
        connected: 'Bağlı',
        connectedDesc:
          'Açık, {{name}} ağına katılır. Kapalı, bağlantıyı keser ve belleği boşaltmak için {{name}} hizmetini durdurur.',
        connectAtBoot: 'Açılışta bağlan',
        connectAtBootDesc: 'KVM açıldığında {{name}} ağına katıl.',
        restartService: 'Hizmeti yeniden başlat',
        needsLogin: 'Giriş gerekli',
        error: 'Hata',
        thisDevice: 'Bu cihaz',
        memoryOf: '{{used}} (eklenti sınırı {{limit}})',
        memoryPressed: 'Eklentilerin bellek grubu sınırına yakın. Swap açmak yardımcı olabilir.',
        peersSummary: 'Eşler: {{total}} eşten {{online}} çevrimiçi',
        peersSummaryIdle: 'Eşler: {{online}} çevrimiçi, {{idle}} isteğe bağlı, toplam {{total}}',
        showOffline: 'Çevrimdışıları göster ({{offline}})',
        hideOffline: 'Çevrimdışıları gizle',
        blocked:
          '{{other}} açık veya açılışta bağlanıyor. Aynı anda yalnızca bir VPN çalışabilir: önce {{other}} için Bağlı ve Açılışta bağlan seçeneklerini kapatın.',
        deviceName: 'Cihaz adı',
        deviceIP: "Cihaz IP'si",
        account: 'Hesap',
        version: 'Sürüm',
        uptime: 'Çalışma süresi',
        noPeers: 'Henüz eş yok.',
        online: 'Çevrimiçi',
        offline: 'Çevrimdışı',
        idle: 'İsteğe bağlı: NetBird, trafik gerektirdiğinde bağlanır',
        memory: 'Bellek',
        uninstall: '{{name}} kaldır',
        uninstallDesc: '{{name}} kaldırılsın mı? Oturum bilgisi kartta kalır.',
        copy: 'Kopyala',
        copied: 'Bağlantı kopyalandı',
        copyFailed: 'Bağlantı kopyalanamadı. Seçip elle kopyalayın.',
        open: 'Aç',
        checkAgain: 'Yeniden denetle',
        notSignedIn: 'Henüz oturum açılmadı. Bağlantıdan oturum açmayı bitirip yeniden denetleyin.',
        checkFailed: 'Oturum durumu denetlenemedi',
        loginWaiting: 'Bu sayfa birkaç saniyede bir denetler ve oturum açtığınızda devam eder.',
        uninstallFailed: 'Kaldırma başarısız',
        loginFailed: 'Oturum açılamadı'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'İndir',
        package: 'yükleme paketi',
        unzip: 'sıkışmış dosyayı açın',
        notLogin: 'Cihaz bağlı değil. Lütfen giriş yapıp cihazınızı hesabınıza bağlayın.',
        urlPeriod: 'Adres sadece 10 ndakika boyunca geçerlidir',
        login: 'Giriş yap',
        logout: 'Çıkış yap',
        logoutDesc: 'Çıkış yapmak istediğinizden emin misiniz?',
        manualIntro: 'Ya da SSH üzerinden elle kurun:',
        copyBinaries:
          "tailscale ve tailscaled dosyalarını IronKVM'deki {{dir}} dizinine kopyalayın",
        linksFile: 'Aynı dizinde, şu iki satırı içeren links adlı bir dosya oluşturun:',
        rebootRefresh: "IronKVM'yi yeniden başlatın, ardından bu sayfayı yenileyin"
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Bu cihaz henüz bir NetBird ağına katılmadı. Bir kurulum anahtarıyla katılın veya SSO ile giriş yapın.',
        setupKey: 'Kurulum anahtarı',
        setupKeyPlaceholder: 'NetBird panelinden bir kurulum anahtarı yapıştırın',
        join: 'Katıl',
        or: 'veya',
        sso: 'SSO ile giriş yap',
        urlPeriod: 'Bu adres 10 dakika boyunca geçerlidir',
        logout: 'Kaydı sil',
        logoutDesc:
          'Kaydı silmek bu eşi NetBird hesabınızdan kaldırır ve buradaki yapılandırmasını siler. Yeniden katılmak için bir kurulum anahtarı veya SSO girişi gerekir ve eş yeni bir IP alabilir. Devam edilsin mi?',
        joinFailed: 'Ağa katılınamadı'
      },
      update: {
        title: 'Güncelleştirmeleri kontrol et',
        queryFailed: 'Sürüm bilgisi alınamadı',
        updateFailed: 'Güncelleme başarısız oldu. Lütfen tekrar deneyin.',
        isLatest: 'En yeni sürüme sahipsiniz.',
        available: 'Bir güncelleme indirilebilir. Şimdi güncellemek istediğinizden emin misiniz?',
        updating: 'Güncelleme başlatıldı. Lütfen bekleyin...',
        confirm: 'Onayla',
        cancel: 'İptal',
        preview: 'Ön İzleme Güncellemeleri',
        previewDesc: 'En son geliştirmelere ve özelliklere erken erişin',
        previewTip:
          'Ön izleme güncellemelerinin tamamlanmamış olduğunu ve sorunlara sebep olabileceğini unutmayın!',
        customServer: {
          title: 'Özel güncelleme sunucusu',
          desc: 'Belirtilen sunucudaki çevrimiçi güncellemeleri denetleyin ve indirin',
          invalidUrl:
            'Sorgu, parça tanımlayıcısı veya latest.json içermeyen geçerli bir HTTP ya da HTTPS sunucu dizini girin.',
          loadFailed: 'Güncelleme sunucusu yapılandırması yüklenemedi.',
          saveFailed: 'Güncelleme sunucusu yapılandırması kaydedilemedi.',
          saved: 'Güncelleme sunucusu yapılandırması kaydedildi.',
          save: 'Kaydet',
          confirmTitle: 'Özel bir güncelleme sunucusu kullanılsın mı?',
          confirmDesc:
            'SHA-512 yalnızca paketin bu sunucunun sağladığı bildirimle eşleştiğini doğrular. Paketin resmi bir IronKVM sürümü olduğunu kanıtlamaz. Hatalı veya kötü amaçlı bir sunucu cihazı kullanılamaz hâle getirebilir, veri kaybına yol açabilir ya da sistem güvenliğini tehlikeye atabilir.',
          confirm: 'Yine de kullan',
          useSipeed: 'Resmi Sipeed sunucusunu kullan',
          previewDisabled:
            'Özel bir güncelleme sunucusu etkinken önizleme güncellemeleri kullanılamaz.'
        },
        offline: {
          chooseFile: 'Dosya seç',
          installing: 'Yükleme tamamlandı. Kuruluyor...',
          noFile: 'Dosya seçilmedi',
          title: 'Çevrimdışı Güncellemeler',
          desc: 'Yerel kurulum paketi aracılığıyla güncelleme',
          upload: 'Yükle',
          checksumPlaceholder: 'SHA-256 sağlama toplamı (isteğe bağlı)',
          invalidChecksum: 'SHA-256 sağlama toplamı 64 onaltılık karakter içermelidir.',
          checksumMismatch: 'SHA-256 doğrulaması başarısız oldu. Paket bozulmuş olabilir.',
          invalidName: 'Geçersiz dosya adı biçimi. Lütfen GitHub sürümlerinden indirin.',
          updateFailed: 'Güncelleme başarısız oldu. Lütfen tekrar deneyin.'
        },
        updateTo: '{{version}} sürümüne güncelle',
        updateConfirmDesc:
          'Cihaz güncellemeyi kurar ve sunucusunu yeniden başlatır. Sunucu geri geldiğinde bu sayfa yeniden yüklenir.',
        releaseNotes: 'Sürüm notları'
      },
      account: {
        title: 'Hesap',
        webAccount: 'Web Hesap Adı',
        role: 'Rol',
        roles: { admin: 'Yönetici', user: 'Kullanıcı' },
        password: 'Şifre',
        updateBtn: 'Değiştir',
        logoutBtn: 'Çıkış  yap',
        logoutDesc: 'Çıkış yapmak istediğinizden emin misiniz?',
        okBtn: 'Evet',
        cancelBtn: 'Hayır',
        users: {
          title: 'Kullanıcılar',
          create: 'Kullanıcı Oluştur',
          enabled: 'Etkin',
          disabled: 'Devre dışı',
          deviceOwner: 'Cihaz sahibi',
          resetPassword: 'Şifreyi Sıfırla',
          delete: 'Sil',
          deleteConfirm: 'Bu kullanıcı silinsin ve tüm oturumları iptal edilsin mi?',
          created: 'Kullanıcı oluşturuldu',
          deleted: 'Kullanıcı silindi',
          passwordUpdated: 'Şifre güncellendi',
          loadFailed: 'Kullanıcılar yüklenemedi',
          saveFailed: 'Kullanıcı kaydedilemedi',
          deleteFailed: 'Kullanıcı silinemedi'
        }
      },
      apiKeys: {
        mcpNote: "Bu anahtarlar MCP için çalışmaz; MCP'nin MCP sayfasında kendi anahtarı vardır.",
        metricsUrl: "Metrik URL'si",
        monitoring: 'İzleme',
        monitoringDesc:
          'Prometheus metrikleri bu sayfadaki bir API anahtarıyla, Bearer token olarak göndererek okur. Her rol okuyabilir.',
        scrapeConfig: 'Prometheus scrape yapılandırması',
        title: 'API Anahtarları',
        description:
          'Bir anahtar, sahibi adına o kullanıcının rolüyle çalışır. Metrikler ve API için Authorization: Bearer <key> olarak, Redfish için X-Auth-Token olarak gönderin.',
        name: 'Ad',
        namePlaceholder: 'Anahtarın amacı, örneğin prometheus',
        nameRequired: 'Anahtara bir ad verin',
        nameTooLong: 'Ad en fazla 64 karakter olabilir',
        unnamed: '(adsız)',
        create: 'Anahtar Oluştur',
        created: 'Oluşturulma',
        owner: 'Sahip',
        empty: 'API anahtarı yok',
        newKeyTitle: 'Yeni API anahtarınız',
        newKeyWarning:
          'Anahtarı şimdi kopyalayın. Saklanmaz ve tekrar gösterilemez. Kaybederseniz iptal edip yenisini oluşturun.',
        copy: 'Kopyala',
        copied: 'Kopyalandı',
        copyFailed: 'Kopyalama başarısız oldu. Elle kopyalayın.',
        done: 'Tamam',
        revoke: 'İptal et',
        revokeConfirmTitle: 'Bu API anahtarı iptal edilsin mi?',
        revokeConfirmDesc: '"{{name}}" anahtarını kullanan her şey hemen çalışmayı durdurur.',
        revoked: 'API anahtarı iptal edildi',
        loadFailed: 'API anahtarları yüklenemedi',
        createFailed: 'API anahtarı oluşturulamadı',
        revokeFailed: 'API anahtarı iptal edilemedi',
        cancelBtn: 'Vazgeç'
      }
    },
    picoclaw: {
      title: 'PicoClaw Asistan',
      empty: 'Paneli açın ve başlamak için bir görevi başlatın.',
      inputPlaceholder: "PicoClaw'nin ne yapmasını istediğinizi açıklayın",
      newConversation: 'Yeni görüşme',
      processing: 'İşleniyor...',
      agent: {
        defaultTitle: 'Genel Asistan',
        defaultDescription: 'Genel sohbet, arama ve çalışma alanı yardımı.',
        kvmTitle: 'Uzaktan Kontrol',
        kvmDescription: 'Uzak ana bilgisayarı IronKVM aracılığıyla çalıştırın.',
        switched: 'Temsilci rolü değiştirildi',
        switchFailed: 'Temsilci rolü değiştirilemedi'
      },
      send: 'Gönder',
      cancel: 'İptal',
      status: {
        connecting: 'Ağ geçidine bağlanılıyor...',
        connected: 'PicoClaw oturumu bağlandı',
        disconnected: 'PicoClaw oturumu bağlantısı kesildi',
        stopped: 'Durdurma isteği gönderildi',
        runtimeStarted: 'PicoClaw runtime başlatıldı',
        runtimeStartFailed: 'PicoClaw runtime başlatılamadı',
        runtimeStopped: 'PicoClaw runtime durduruldu',
        runtimeStopFailed: 'PicoClaw runtime durdurulamadı',
        controlSwitchedToMCP: 'Kontrol harici MCP hizmetine geçirildi'
      },
      connection: {
        runtime: {
          checking: 'Kontrol ediliyor',
          restoring: 'PicoClaw geri yükleniyor',
          ready: 'Runtime hazır',
          stopped: 'Runtime durduruldu',
          blockedByMCP: 'Harici MCP kontrolü etkin',
          readyBlockedByMCP:
            'Runtime çalışıyor, ancak cihaz girişini şu anda harici MCP kontrol ediyor.',
          readyWithoutControl:
            "Runtime çalışıyor. Yeniden bağlanmadan önce PicoClaw'a cihaz kontrolü verin.",
          unavailable: 'Runtime mevcut değil',
          configError: 'Yapılandırma hatası'
        },
        transport: {
          connecting: 'Bağlanıyor',
          connected: 'Bağlandı',
          disconnected: 'Bağlantı kesildi',
          reconnect: 'Yeniden bağlan',
          reconnectDescription: 'Çalışan PicoClaw oturumuna yeniden bağlan.',
          reconnectBlocked: "PicoClaw'ın yeniden bağlanmak için cihaz kontrolüne ihtiyacı var."
        },
        run: {
          idle: 'Boşta',
          busy: 'Meşgul'
        }
      },
      message: {
        toolAction: 'Eylem',
        observation: 'Gözlem',
        screenshot: 'Ekran Görüntüsü'
      },
      overlay: {
        locked: 'PicoClaw cihazı kontrol ediyor. Manuel giriş duraklatıldı.'
      },
      control: {
        picoclaw: 'Cihaz kontrolü: PicoClaw',
        picoclawDescription:
          'PicoClaw klavye ve fare girişi gönderebilir. Elle giriş duraklayabilir.',
        mcp: 'Cihaz kontrolü: harici MCP',
        mcpDescription: 'Harici MCP cihaza yazabilir. PicoClaw girişi devralmaz.',
        off: 'Cihaz kontrolü: kapalı',
        offDescription:
          'Yapay zekâ klavye veya fare girişi göndermez. Elle kontrol kullanılabilir kalır.',
        transitioning: 'Cihaz kontrolü: geçiş yapılıyor',
        transitioningDescription: 'Cihaz kontrolü eşitleniyor. Lütfen bekleyin.',
        grant: 'Kontrol ver',
        release: 'Bırak',
        releasing: 'Bırakılıyor...',
        switching: 'Geçiş yapılıyor...',
        releasingLabel: 'Cihaz kontrolü: bırakılıyor',
        releasingDescription: 'Cihaz kontrolü geri veriliyor. PicoClaw süren yazmaları durdurdu.',
        granted: 'PicoClaw kontrolü verildi',
        released: 'PicoClaw kontrolü bırakıldı',
        grantFailed: 'PicoClaw kontrolü verilemedi',
        releaseFailed: 'PicoClaw kontrolü bırakılamadı',
        grantConfirmTitle: "Cihaz kontrolü PicoClaw'a geçirilsin mi?",
        grantConfirmDesc: 'Harici MCP cihaz yazmaları kesintiye uğrayacak.'
      },
      install: {
        install: 'Yükle PicoClaw',
        installing: 'PicoClaw yükleniyor',
        success: 'PicoClaw başarıyla yüklendi',
        failed: 'PicoClaw yüklenemedi',
        uninstalling: 'Runtime kaldırılıyor...',
        uninstalled: 'Runtime başarıyla kaldırıldı.',
        uninstallFailed: 'Kaldırma başarısız oldu.',
        requiredTitle: 'PicoClaw kurulu değil',
        requiredDescription: "PicoClaw runtime'ı başlatmadan önce PicoClaw'ı yükleyin.",
        progressDescription: 'PicoClaw kuruluyor.',
        stages: {
          preparing: 'Hazırlanıyor',
          downloading: 'İndiriliyor',
          extracting: 'Çıkarılıyor',
          verifying: 'Doğrulanıyor',
          installing: 'Yükleniyor',
          installed: 'Yüklendi',
          install_timeout: 'Zaman Aşımı',
          install_failed: 'Başarısız'
        }
      },
      model: {
        requiredTitle: 'Model yapılandırması gerekli',
        requiredDescription: 'PicoClaw sohbetini kullanmadan önce PicoClaw modelini yapılandırın.',
        docsTitle: 'Yapılandırma Kılavuzu',
        docsDesc: 'Desteklenen modeller ve protokoller',
        menuLabel: 'Modeli yapılandır',
        modelIdentifier: 'Model Tanımlayıcı',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API Anahtarı',
        apiKeyPlaceholder: 'Model API anahtarını girin',
        apiKeyOptionalPlaceholder: 'ollama, lmstudio veya vllm için gerekmez',
        save: 'Kaydet',
        saving: 'Kaydediliyor',
        saved: 'Model yapılandırması kaydedildi',
        saveFailed: 'Model yapılandırması kaydedilemedi',
        invalid: 'Model tanımlayıcı, API Base URL ve API anahtarı gereklidir',
        invalidNoKey: 'Model tanımlayıcı ve API Base URL gereklidir'
      },
      uninstall: {
        menuLabel: 'Kaldırma',
        confirmTitle: 'Kaldırma PicoClaw',
        confirmContent:
          "PicoClaw'yi kaldırmak istediğinizden emin misiniz? Bu, yürütülebilir dosyayı ve tüm yapılandırma dosyalarını siler.",
        confirmOk: 'Kaldırma',
        confirmCancel: 'İptal'
      },
      history: {
        title: 'Geçmiş',
        loading: 'Oturumlar yükleniyor...',
        emptyTitle: 'Henüz geçmiş yok',
        emptyDescription: 'Önceki PicoClaw oturumlar burada görünecek.',
        loadFailed: 'Oturum geçmişi yüklenemedi',
        deleteFailed: 'Oturum silinemedi',
        deleteConfirmTitle: 'Oturumu sil',
        deleteConfirmContent: '"{{title}}" silmek istediğinizden emin misiniz?',
        deleteConfirmOk: 'Sil',
        deleteConfirmCancel: 'İptal',
        messageCount_one: '{{count}} mesaj',
        messageCount_other: '{{count}} mesaj',
        messageCount: '{{count}} mesaj'
      },
      config: {
        startRuntime: "PicoClaw'ı Başlat",
        stopRuntime: "PicoClaw'ı Durdur"
      },
      start: {
        enableConfirmTitle: "Kontrol PicoClaw'a geçirilsin mi?",
        enableConfirmDesc: "PicoClaw'ı başlatmak harici MCP hizmetini devre dışı bırakır.",
        enableConfirmOk: "PicoClaw'ı Başlat",
        enableConfirmCancel: 'İptal',
        title: "PicoClaw'ı Başlat",
        description: "PicoClaw yardımcısını kullanmaya başlamak için runtime'ı başlatın.",
        switchFromMCP: "PicoClaw'a geç ve başlat",
        takeoverAndStart: 'Devral ve başlat'
      }
    },
    error: {
      title: 'Bir hata oldu!',
      refresh: 'Yenile',
      panel: 'Sayfanın bu bölümü çalışmayı durdurdu',
      retry: 'Tekrar dene'
    },
    fullscreen: {
      toggle: 'Tam ekrana geç'
    },
    input: {
      disconnected: 'Klavye ve fare bağlı değil',
      disconnectedTls:
        'Tarayıcı, klavye ve fareyi taşıyan güvenli bağlantıyı sormadan reddetti. Bu cihazın oluşturduğu sertifikaya henüz güvenilmiyor. Bu adresi yeni bir sekmede açın, sertifikayı kabul edin ve ardından sayfayı yeniden yükleyin. Kalıcı çözüm sertifikayı yüklemektir.',
      disconnectedNever:
        'Klavye ve fareyi taşıyan bağlantı açılamadı. Sayfanın geri kalanı bu bağlantıyı kullanmadığı için çalışıyor. Sizinle cihaz arasında hiçbir şeyin onu engellemediğinden emin olun.',
      disconnectedDropped:
        'Klavye ve fareyi taşıyan bağlantı koptu ve geri gelmedi. Yeniden başlatmanın ardından kendiliğinden yeniden bağlanır; bu durum sürerse sayfayı yeniden yükleyin.',
      hidDisabled: 'Bu cihazda HID kapalı (/boot/disable_hid).',
      keyFailed: 'Tuş gönderilemedi.'
    },
    speaker: {
      title: 'Hoparlör',
      unmute: 'Sesi aç',
      mute: 'Sesi kapat',
      hostIdle: 'Ana makine ses göndermiyor',
      hostIdleHint: "Ana makinede bir şey çalın veya ses çıkışı olarak KVM'yi seçin.",
      host: 'Ana bilgisayar sesi',
      room: 'Oda mikrofonu',
      roomListen: 'Odayı dinle',
      roomVolume: 'Ses düzeyi',
      roomHint: 'Açıkken izleyen herkes görür.',
      roomLive: 'Mikrofon açık',
      roomLiveBy: 'Oda mikrofonu açık: {{names}}'
    },
    upstream: {
      check: 'Güncellemeleri denetle',
      updateTo: '{{version}} sürümüne güncelle',
      confirm: '{{name}}, {{version}} sürümüne güncellensin mi?',
      confirmDesc:
        "Yeni sürüm GitHub'dan indirilir ve yayımladığı sağlama toplamlarıyla doğrulanır. Bir şey başarısız olursa mevcut sürüm kalır.",
      ok: 'Güncelle',
      upToDate: 'Güncel',
      builtIn: 'yerleşik',
      checkFailed: 'Güncellemeler denetlenemedi: {{error}}',
      unverifiable: '{{version}} sürümü sunulmuyor: {{reason}}',
      inUse: 'Şu anda güncellenemiyor: {{reason}}',
      running: '{{version}} sürümüne güncelleniyor...',
      done: '{{name}}, {{version}} sürümüne güncellendi',
      failed: 'Son güncelleme başarısız oldu: {{error}}'
    },
    menu: {
      mediaAdd: 'Kalıp ekle',
      mediaMoreOptions: 'Diğer seçenekler',
      mediaSettings: 'Medya ayarları',
      collapse: 'Menüyü küçült',
      expand: 'Menüyü genişlet',
      more: 'Daha fazla',
      media: 'Medya',
      tools: 'Araçlar',
      text: 'Metin',
      advanced: 'Gelişmiş',
      mediaMounted: 'Bağlı',
      mediaLibrary: 'Kitaplık',
      textToHost: 'Ana makineye',
      textFromHost: 'Ana makineden'
    },
    ion: {
      checking: 'Akış başlatılmadan önce video belleği kontrol ediliyor...',
      warn: 'Video belleği az. Tek bir sunucu yeniden başlatması belleği tüketir. Uygun olduğunda yeniden başlatın.',
      criticalTitle: 'Akışı başlatmak için yeterli video belleği yok',
      criticalBody:
        "Videoyu başlatmak ayrılmış belleği tüketir ve sunucuyu durdurur. Güç kontrolü ve yeniden başlatma dahil diğer tüm işlevler çalışmaya devam eder. Bu belleği yalnızca IronKVM'i yeniden başlatmak geri kazandırır.",
      criticalContinue: 'Videoyu yine de başlat',
      criticalReboot: "IronKVM'i Yeniden Başlat",
      criticalRebooting: 'Yeniden başlatılıyor...'
    }
  }
};

export default tr;
