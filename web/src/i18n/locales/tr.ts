const tr = {
  translation: {
    head: {
      desktop: 'Uzak masaüstü',
      login: 'Giriş',
      changePassword: 'Şifreyi değiştir',
      terminal: 'Uçbirim',
      wifi: 'Wi-Fi'
    },
    auth: {
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
          'Şifreleri sıfırlamak için NanoKVM üzerinde bulunan BOOT tuşuna 10 saniye boyunca basılı tutun.',
        reset2: 'Ayrıntılı adımlar için dökümana göz atın:',
        reset3: 'Arayüz varsayılan hesap:',
        reset4: 'Güvenli Kabuk Bağlantısı (SSH) varsayılan hesap:',
        change1: 'Bu işlem şu şifreleri değiştiricektir:',
        change2: 'Arayüz giriş şifresi',
        change3: 'Sistem yöneticisi şifresi (Güvenli Kabuk Bağlantısı (SSH) giriş şifresi)',
        change4:
          'Şifreleri sıfırlamak için NanoKVM üzerinde bulunan BOOT tuşuna 10 saniye boyunca basılı tutun.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'NanoKVM için Wi-Fi ayarlarını ayarlayın',
      success: "NanoKVM'in bağlantı durumunu kontrol edin ve yeni IP adresini ziyaret edin.",
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
      }
    },
    screen: {
      scale: 'Ölçek',
      title: 'Ekran',
      video: 'Görüntü modu',
      videoDirectTips: 'kullanmak için "Ayarlar > Cihaz" HTTPS aktif edin',
      resolution: 'Çözünürlük',
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
      qualityLossless: 'Kayıpsız',
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
      }
    },
    keyboard: {
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
      scrollUp: 'Yukarı kaydır',
      scrollDown: 'Aşağı kaydır',
      speed: 'Kaydırma tekerleği hızı',
      fast: 'Hızlı',
      slow: 'Yavaş',
      requestPointer: 'Bağıl fare modu kullanılıyor. Masaüstüne tıklayarak imleç elde edinin.',
      resetHid: 'HID’yi sıfırla',
      hidOnly: {
        title: 'Yalnızca HID modu',
        desc: 'Fare ve klavye yanıt vermeyi durdurursa ve HID sıfırlama yardımcı olmazsa, NanoKVM ile cihaz arasında bir uyumluluk sorunu olabilir. Daha iyi uyumluluk için yalnızca HID modunu etkinleştirmeyi deneyin.',
        tip1: 'Yalnızca HID modunu etkinleştirmek sanal U-disk’i ve sanal ağı ayırır',
        tip2: 'Yalnızca HID modunda imaj bağlama devre dışıdır',
        rebuild: 'Mod değiştirmek USB bağlantısını yeniden kurar. NanoKVM yeniden başlamaz',
        enable: 'Yalnızca HID modunu etkinleştir',
        disable: 'Yalnızca HID modunu devre dışı bırak'
      }
    },
    image: {
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
        usb1: "NanoKVM'i bilgisayarınıza USB ile bağlayın.",
        usb2: 'Sanal diskin bağlı olduğundan emin olun (Ayarlar - Sanal Disk).',
        usb3: 'Sanal diski bilgisayarınızda açın ve disk imajı dosyanızı sanal diskin kök dizinine kopyalayın.',
        scp1: 'NanoKVM ve bilgisayarınızın aynı yerel ağda bulunduğundan emin olun.',
        scp2: "Bilgisayarınızda uçbirimi açın ve disk imajı dosyanını SCP komudunu kullanarak NanoKVM'in /data dizinine yükleyin.",
        scp3: 'Örnek: scp senin-disk-imajı-dizinin root@senin-nanokvm-ip:/data',
        tfCard: 'micro SD kart',
        tf1: 'Bu yöntem Linux sistemlerde desteklenmektedir.',
        tf2: "NanoKVM'den micro SD kartı çıkartın(TAM sürüm için öncelikle kutuyu sökün).",
        tf3: 'micro SD kartı kart okuyucusuna takın ve bilgisayarınıza bağlayın.',
        tf4: 'Disk imajı dosyanını micro SD kartın /data dizinine kopyalayın.',
        tf5: "micro SD kartı NanoKVM'e geri yerleştirin."
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
      close: 'Kapat'
    },
    terminal: {
      title: 'Uçbirim',
      nanokvm: 'NanoKVM Uçbirimi',
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
      title: 'Ağ Üzerinden Uyandırma (WOL)',
      sending: 'Komut gönderiliyor...',
      sent: 'Komut gönderildi',
      input: 'MAC adresi girin',
      ok: 'Tamam'
    },
    download: {
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
      bootMenuDesc: "Sanal CD için netboot.xyz ISO'sunu sağlama toplamı doğrulanmış olarak indirin"
    },
    power: {
      title: 'Güç',
      showConfirm: 'Doğrulama',
      showConfirmTip: 'Güç ile ilgili işlemler fazladan doğrulama gerektirir',
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
      ledConnectedFailed: "Güç LED'i ayarı kaydedilemedi"
    },
    settings: {
      title: 'Ayarlar',
      nav: {
        general: 'Genel',
        device: 'Cihaz',
        network: 'Ağ',
        remote: 'Uzaktan erişim',
        boot: 'Önyükleme',
        locked: 'Bir işlem sürüyor. Bitene kadar diğer sayfalar ve kapatma kullanılamaz.',
        vpnProvider: 'VPN sağlayıcısı'
      },
      mcp: {
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
        cancelBtn: 'İptal'
      },
      redfish: {
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
      vnc: {
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
        actionReset: 'Reset',
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
        failed: 'Watchdog işlemi başarısız oldu'
      },
      netboot: {
        title: 'Ağdan önyükleme',
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
        title: 'NanoKVM Hakkında',
        information: 'Bilgi',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Uygulama sürümü',
        applicationTip: 'NanoKVM web uygulaması sürümü',
        image: 'İmaj Sürümü',
        imageTip: 'NanoKVM sistem imajı sürümü',
        kernel: 'Çekirdek Sürümü',
        kernelTip: 'Şu anda çalışan Linux çekirdeğinin sürümü',
        deviceKey: 'Cihaz Anahtarı',
        videoMemory: 'Video Belleği',
        videoMemoryTip: 'Video yakalama için ayrılmış bellek. Sistemin geri kalanıyla paylaşılmaz.',
        videoMemoryGenerations_one: '{{count}} önceki NanoKVM oturumu video belleğini tutuyor',
        videoMemoryGenerations_other: '{{count}} önceki NanoKVM oturumu video belleğini tutuyor',
        videoMemoryReboot: 'Geri kazanmak için yeniden başlatın.',
        community: 'Topluluk',
        hostname: 'Ana makine adı',
        hostnameUpdated: 'Hostname güncellendi. Uygulamak için yeniden başlatın.',
        ipType: {
          Wired: 'Kablolu bağlantı',
          Wireless: 'Kablosuz bağlantı',
          Other: 'Diğer'
        }
      },
      appearance: {
        title: 'Görünüm',
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
        ssh: {
          description: 'Güvenli Kabuk Bağlantısı (SSH) aktif et',
          tip: 'Aktifleştirmeden önce güçlü bir şifreye sahip olduğunuzdan emin olun (Hesap - Şifremi Değiştir)'
        },
        advanced: 'Gelişmiş Ayarlar',
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
          tip: 'Bu özelliği aktifleştirmek micro SD kartınızın ömrünü kısaltabilir!'
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
          description:
            'USB kablosu üzerinden uzak ana bilgisayarla özel bir ağ bağlantısı. Ana bilgisayar ağ geçidi ve DNS olmadan bir adres alır, bu yüzden NanoKVM üzerinden yerel ağınıza ulaşamaz.',
          off: 'Kapalı',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (NCM desteği olmayan ana bilgisayarlar için)',
          rndis: 'RNDIS (artık sunulmuyor)',
          rndisNote: 'Bu bağlantı artık sunulmayan RNDIS kullanıyor. NCM veya ECM seçin.',
          subnet: 'Alt ağ',
          subnetDesc:
            '/24 ile /30 arasında özel bir IPv4 ağı. NanoKVM ilk adresi, ana bilgisayar ikinci adresi alır.',
          addresses: 'NanoKVM: {{board}}, ana bilgisayar: {{host}}',
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
          "Ağa erişilemediğinde bu NanoKVM'e giriş yapabilmek için uzak ana bilgisayara bir USB seri port sunar",
        consoleTip:
          'Uzak ana bilgisayarı kontrol eden herkes bu NanoKVM için bir giriş istemi görür. Etkinleştirmeden önce güçlü bir şifre belirleyin (Hesap - Şifremi Değiştir).',
        endpoints: {
          title: 'USB uç noktaları',
          used: '{{used}} / {{total}} kullanımda',
          cost: '{{cost}} kullanıyor',
          needs: '{{cost}} gerekli',
          full: 'Yeterli USB uç noktası yok. Önce başka bir şeyi kapatın.',
          inactive:
            'Açık, ancak çalışmıyor: USB denetleyicisinin uç noktaları tükendi. Başka bir aygıtı kapatın, bu aygıt hemen başlar.',
          explain:
            'USB denetleyicisinin sabit sayıda giriş uç noktası vardır ve bu sayaç onları sayar. Sığabilecek olandan fazla aygıt etkinse klavye ve fare korunur, diğerleri kapatılır.',
          error: 'Cihaza ulaşılamadı. Tekrar deneyin.',
          fitTogether: 'Birlikte sığanlar: {{sets}}'
        },
        reboot: 'Yeniden Başlat',
        rebootDesc: "NanoKVM'i yeniden başlatmak istediğinizden emin misiniz?",
        okBtn: 'Evet',
        cancelBtn: 'Hayır'
      },
      network: {
        title: 'Ağ',
        wifi: {
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
          waitingHttp: "http'ye geri dönülüyor. Sayfa kendiliğinden açılmazsa yeniden yükleyin."
        },
        ethernet: {
          title: 'IP Adresi',
          description: 'NanoKVM cihazının kablolu ağdaki adresini nasıl aldığını yapılandırın',
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
          applyTitle: 'NanoKVM adresi değiştirilsin mi?',
          applyWarning:
            'Bu sayfayla bağlantı kesilecek. NanoKVM yeni adresi uygular ve ona bu adresten ulaşmanız için {{seconds}} saniye bekler. Ona ulaşmak değişikliği korur. Ona hiçbir şey ulaşmazsa NanoKVM önceki ayarları geri yükler.',
          applyConfirm: 'Uygula',
          applyCancel: 'İptal',
          applyFailed: 'Adres uygulanamadı',
          trialTitle: 'Onay bekleniyor',
          trialDhcp: 'NanoKVM, DHCP üzerinden adres istiyor.',
          trialStatic: 'NanoKVM şimdi {{address}} adresinde.',
          trialInstruction:
            "NanoKVM'i yeni adresinde açın ve isterse oturum açın. Ona orada ulaşmak değişikliği korur. {{seconds}} saniye içinde NanoKVM'e hiçbir şey ulaşmazsa önceki ayarlar geri yüklenir.",
          trialOpen: 'Yeni adresi aç',
          trialKeep: 'Bu ayarları koru',
          trialKept: 'Yeni adres kaydedildi',
          trialKeepFailed: 'Ayarlar korunamadı',
          trialGone: 'Değişiklik zaten geri alındı. Tekrar deneyin.',
          unsaved: 'Kaydedilmemiş değişiklikler'
        },
        dns: {
          title: 'DNS',
          description: 'NanoKVM için DNS sunucularını yapılandır',
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
        }
      },
      vpn: {
        loading: 'Yükleniyor...',
        okBtn: 'Evet',
        cancelBtn: 'Hayır',
        restart: '{{name}} yeniden başlatılsın mı?',
        stop: '{{name}} durdurulsun mu?',
        stopDesc:
          'Arka plan hizmeti şimdi durur. Açılışta başlatma ayrı bir anahtardır ve olduğu gibi kalır.',
        update: '{{name}}, {{version}} sürümüne güncellensin mi?',
        updateDesc: 'Arka plan hizmeti çalışıyorsa yeniden başlar. Oturum bilgisi korunur.',
        notInstall: '{{name}} yüklü değil.',
        install: 'Yükle',
        installing: 'Yükleniyor',
        installFailed: 'Yükleme başarısız oldu',
        retry: 'Tekrar dene',
        notRunning: '{{name}} çalışmıyor. Devam etmek için başlatın.',
        run: 'Başlat',
        boot: 'Açılışta başlat',
        bootDesc: 'KVM açılırken {{name}} başlatılsın.',
        enable: '{{name}} etkinleştir',
        control: 'Kontrol sunucusu',
        connected: 'Bağlı',
        disconnected: 'Bağlı değil',
        deviceName: 'Cihaz adı',
        deviceIP: "Cihaz IP'si",
        account: 'Hesap',
        version: 'Sürüm',
        uptime: 'Çalışma süresi',
        peers: 'Eşler',
        noPeers: 'Henüz eş yok.',
        online: 'Çevrimiçi',
        offline: 'Çevrimdışı',
        memory: 'Bellek',
        daemonRss: 'Hizmet',
        group: 'Eklentiler grubu',
        high: '{{size}} üzerinde yavaşlatılır',
        max: '{{size}} üzerinde çekirdek tarafından durdurulur',
        noGroup: 'Bu kartta eklenti bellek grubu yok.',
        uninstall: '{{name}} kaldır',
        uninstallDesc: '{{name}} kaldırılsın mı? Oturum bilgisi kartta kalır.',
        blocked:
          '{{other}} çalışıyor veya açılışta başlıyor. Aynı anda yalnızca bir VPN çalışabilir: önce {{other}} hizmetini durdurun ve açılışta başlatmayı kapatın.',
        swap: {
          title: 'Swap belleği',
          tip: 'Hizmetin belleği yetmezse swap belleğini etkinleştirmeyi deneyin. Bu, swap dosyasının boyutunu varsayılan olarak 256MB yapar; boyut "Ayarlar > Cihaz" bölümünden değiştirilebilir.'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Lütfen sayfayı yenileyin ve tekrar deneyin, ya da manuel indirin',
        download: 'İndir',
        package: 'yükleme paketi',
        unzip: 'sıkışmış dosyayı açın',
        upTailscale: "tailscale dosyasını NanoKVM'in /usr/bin dizinine yükleyin",
        upTailscaled: "tailscaled dosyasını NanoKVM'in /usr/sbin dizinine yükleyin",
        refresh: 'İçinde bulunduğunuz sayfayı yenileyin',
        notLogin: 'Cihaz bağlı değil. Lütfen giriş yapıp cihazınızı hesabınıza bağlayın.',
        urlPeriod: 'Adres sadece 10 ndakika boyunca geçerlidir',
        login: 'Giriş yap',
        loginSuccess: 'Giriş yapıldı',
        logout: 'Çıkış yap',
        logoutDesc: 'Çıkış yapmak istediğinizden emin misiniz?'
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
        loginSuccess: 'Giriş yapıldı',
        logout: 'Kaydı sil',
        logoutDesc:
          'Kaydı silmek bu eşi NetBird hesabınızdan kaldırır ve buradaki yapılandırmasını siler. Yeniden katılmak için bir kurulum anahtarı veya SSO girişi gerekir ve eş yeni bir IP alabilir. Devam edilsin mi?'
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
            'SHA-512 yalnızca paketin bu sunucunun sağladığı bildirimle eşleştiğini doğrular. Paketin resmi bir NanoKVM sürümü olduğunu kanıtlamaz. Hatalı veya kötü amaçlı bir sunucu cihazı kullanılamaz hâle getirebilir, veri kaybına yol açabilir ya da sistem güvenliğini tehlikeye atabilir.',
          confirm: 'Yine de kullan',
          useSipeed: 'Resmi Sipeed sunucusunu kullan',
          previewDisabled:
            'Özel bir güncelleme sunucusu etkinken önizleme güncellemeleri kullanılamaz.'
        },
        offline: {
          title: 'Çevrimdışı Güncellemeler',
          desc: 'Yerel kurulum paketi aracılığıyla güncelleme',
          upload: 'Yükle',
          checksumPlaceholder: 'SHA-256 sağlama toplamı (isteğe bağlı)',
          invalidChecksum: 'SHA-256 sağlama toplamı 64 onaltılık karakter içermelidir.',
          checksumMismatch: 'SHA-256 doğrulaması başarısız oldu. Paket bozulmuş olabilir.',
          invalidName: 'Geçersiz dosya adı biçimi. Lütfen GitHub sürümlerinden indirin.',
          updateFailed: 'Güncelleme başarısız oldu. Lütfen tekrar deneyin.'
        }
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
        kvmDescription: 'Uzak ana bilgisayarı NanoKVM aracılığıyla çalıştırın.',
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
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime hazır',
          stopped: 'Runtime durduruldu',
          blockedByMCP: 'Harici MCP kontrolü etkin',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime mevcut değil',
          configError: 'Yapılandırma hatası'
        },
        transport: {
          connecting: 'Bağlanıyor',
          connected: 'Bağlandı',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
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
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Cihaz kontrolü: harici MCP',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Cihaz kontrolü: kapalı',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Kontrol ver',
        release: 'Bırak',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
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
        progressDescription: 'PicoClaw indiriliyor ve kuruluyor.',
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
        save: 'Kaydet',
        saving: 'Kaydediliyor',
        saved: 'Model yapılandırması kaydedildi',
        saveFailed: 'Model yapılandırması kaydedilemedi',
        invalid: 'Model tanımlayıcı, API Base URL ve API anahtarı gereklidir'
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
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
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
    speaker: { title: 'Hoparlör', unmute: 'Sesi aç', mute: 'Sesi kapat' },
    menu: {
      collapse: 'Menüyü küçült',
      expand: 'Menüyü genişlet',
      more: 'Daha fazla'
    },
    ion: {
      checking: 'Akış başlatılmadan önce video belleği kontrol ediliyor...',
      warn: 'Video belleği az. Tek bir sunucu yeniden başlatması belleği tüketir. Uygun olduğunda yeniden başlatın.',
      criticalTitle: 'Akışı başlatmak için yeterli video belleği yok',
      criticalBody:
        "Videoyu başlatmak ayrılmış belleği tüketir ve sunucuyu durdurur. Güç kontrolü ve yeniden başlatma dahil diğer tüm işlevler çalışmaya devam eder. Bu belleği yalnızca NanoKVM'i yeniden başlatmak geri kazandırır.",
      criticalContinue: 'Videoyu yine de başlat',
      criticalReboot: "NanoKVM'i Yeniden Başlat",
      criticalRebooting: 'Yeniden başlatılıyor...'
    }
  }
};

export default tr;
