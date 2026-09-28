const ja = {
  translation: {
    head: {
      desktop: 'リモートデスクトップ',
      login: 'ログイン',
      changePassword: 'パスワード変更',
      terminal: 'ターミナル',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        'ブラウザーがセッションの保存を拒否しました。以前の HTTPS セッションで残った Cookie は、通常の http 接続では置き換えられません。このアドレスの Cookie を削除するか、プライベートウィンドウを開いてから、もう一度サインインしてください。',
      login: 'ログイン',
      placeholderUsername: 'ユーザー名を入力してください',
      placeholderPassword: 'パスワードを入力してください',
      placeholderCurrentPassword: '現在のパスワード',
      placeholderPassword2: 'パスワードをもう一度入力してください',
      noEmptyUsername: 'ユーザー名は空にできません',
      noEmptyPassword: 'パスワードは空にできません',
      passwordLength: 'パスワードは 8～72 文字で入力してください',
      noAccount:
        'ユーザー情報の取得に失敗しました。ページを更新してもう一度お試しいただくか、パスワードをリセットしてください。',
      invalidUser: 'ユーザー名またはパスワードが正しくありません',
      locked: 'ログインが多すぎます。後でもう一度お試しください。',
      globalLocked: 'システムは保護されています。後でもう一度試してください。',
      error: '不明なエラー',
      invalidCurrentPassword: '現在のパスワードが正しくありません',
      changePassword: 'パスワード変更',
      changePasswordDesc: 'デバイスのセキュリティのために、パスワードを変更してください！',
      differentPassword: 'パスワードが一致しません',
      illegalUsername: 'ユーザー名に不正な文字が含まれています',
      illegalPassword: 'パスワードに不正な文字が含まれています',
      forgetPassword: 'パスワードを忘れた',
      ok: 'OK',
      cancel: 'キャンセル',
      loginButtonText: 'ログイン',
      tips: {
        reset1: 'パスワードをリセットするには、NanoKVM の BOOT ボタンを 10 秒間押し続けます。',
        reset2: '詳細な手順については、次のドキュメントを参照してください：',
        reset3: 'ウェブデフォルトアカウント：',
        reset4: 'SSH デフォルトアカウント：',
        change1: 'この操作により、以下のパスワードも更新されることに注意してください：',
        change2: 'ウェブログインパスワード',
        change3: 'システム root パスワード（SSH ログインパスワード）',
        change4:
          'パスワードを忘れた場合は、NanoKVM の BOOT ボタンを長押ししてパスワードをリセットする必要があります。'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'NanoKVM の Wi-Fi を設定する',
      success: 'NanoKVM のネットワークステータスを確認するにはデバイスにアクセスしてください。',
      failed: '操作に失敗しました。もう一度お試しください。',
      invalidMode:
        '現在のモードではネットワーク設定はサポートされていません。デバイスで Wi-Fi 設定モードを有効にしてください。',
      confirmBtn: 'OK',
      finishBtn: '完了',
      ap: {
        authTitle: '認証が必要です',
        authDescription: '続行するには、AP パスワードを入力してください',
        authFailed: '無効な AP パスワード',
        passPlaceholder: 'AP パスワード',
        verifyBtn: '確認する'
      }
    },
    screen: {
      scale: '倍率',
      title: '画面',
      video: 'ビデオモード',
      videoDirectTips: 'このモードを使用するには「設定 - デバイス」で HTTPS を有効にしてください',
      resolution: '解像度',
      controlRegion: {
        title: 'マウス位置補正',
        description:
          '操作対象のデバイスが 16:9 以外の解像度を使用していて、カーソルの位置が水平方向または垂直方向にずれる場合に使用します。',
        off: 'オフ',
        auto: '自動',
        autoWarning:
          'ユーザーアプリケーションの背景が完全な黒の場合、補正に失敗することがあります。',
        manual: '手動',
        selectedResolution: '選択領域の解像度',
        unused: '未使用',
        originalResolution: '元の解像度',
        selectResolution: '元の解像度を選択',
        addResolution: 'カスタム解像度を追加',
        add: '追加',
        duplicateResolution: 'この解像度はすでに存在します。',
        width: '幅',
        height: '高さ',
        apply: '計算して適用',
        invalidResolution: 'ビデオの準備完了後、有効な元の解像度を入力してください。',
        select: '領域を選択',
        clear: '自動に戻す',
        saveFailed: '入力領域を保存できませんでした。',
        tooSmall: '選択した領域が小さすぎます。',
        previewUnavailable: 'プレビューを利用できません',
        clearConfirm: '黒帯の自動検出に戻しますか？',
        dragHint: 'ドラッグしてリモートデスクトップの領域を選択',
        finish: '完了',
        confirm: '確認',
        cancel: 'キャンセル'
      },
      auto: '自動',
      autoTips:
        '特定の解像度で画面のちらつきやマウスカーソルのずれが発生する場合があります。リモートホストの解像度を調整するか、自動モードを無効にしてください。',
      fps: 'フレームレート',
      customizeFps: 'カスタマイズ',
      quality: '画質',
      qualityLossless: 'ロスレス',
      qualityHigh: '高',
      qualityMedium: '中',
      qualityLow: '低',
      frameDetect: 'フレーム差分検出',
      frameDetectTip:
        'フレーム間の差異を計算し、リモートホストの画面が変更されない場合はビデオストリームの送信を停止します',
      resetHdmi: 'HDMI をリセット',
      mixedH264: {
        title: 'H.264 ストリームの競合',
        description:
          'H.264 Direct と H.264 WebRTC が同時に使用されています。画面のティアリングや映像の破損が発生する可能性があります。H.264 モードは 1 つだけ使用してください。'
      },
      webrtcConnectionFailed: {
        title: 'WebRTC 接続に失敗しました',
        description: 'ネットワーク接続を確認するか、ビデオモードを切り替えてください。'
      },
      captureStatus: {
        hdmiError: 'HDMI 画面エラー',
        unsupportedResolution: '現在の解像度はサポートされていません',
        retrieving: '画面を取得中...',
        changingResolution: '解像度を切り替え中...',
        updateFailed: '現在、画面を更新できません',
        videoError: '映像表示エラー',
        noHdmi: 'HDMI 信号が検出されません',
        unavailable: '現在、画面を表示できません'
      }
    },
    keyboard: {
      title: 'キーボード',
      paste: '貼り付け',
      tips: 'テキストをキー入力としてホストに入力します。ホストで使われているキーボード配列を選んでください。',
      placeholder: '入力してください',
      submit: '送信',
      virtual: '仮想キーボード',
      readClipboard: 'クリップボードから読み取る',
      clipboardPermissionDenied:
        'クリップボードのアクセス許可が拒否されました。ブラウザでクリップボードへのアクセスを許可してください。',
      clipboardReadError: 'クリップボードの読み取りに失敗しました',
      mediaKeys: {
        title: 'メディアキー',
        mute: 'ミュート',
        volumeDown: '音量を下げる',
        volumeUp: '音量を上げる',
        previous: '前のトラック',
        playPause: '再生または一時停止',
        next: '次のトラック',
        stop: '停止'
      },
      pasting: {
        layout: 'ホストのキーボード配列',
        layouts: {
          us: '英語（US）',
          uk: '英語（UK）',
          de: 'ドイツ語',
          fr: 'フランス語',
          es: 'スペイン語',
          it: 'イタリア語',
          ptBr: 'ポルトガル語（ブラジル）',
          se: 'スウェーデン語 / フィンランド語',
          ru: 'ロシア語',
          ja: '日本語',
          ko: '韓国語'
        },
        speed: '入力速度',
        speeds: {
          fast: '速い',
          normal: '標準',
          slow: '遅い'
        },
        estimate: '入力時間: 約 {{duration}}',
        untypeable: 'この配列で入力できない文字: {{count}}',
        untypeableAt: '{{line}} 行 {{column}} 列',
        skipUntypeable: '残りを入力',
        shortcut: '{{shortcut}} でクリップボードの内容をすぐにホストへ入力します。',
        clipboardUnavailable:
          'ブラウザーは HTTPS の場合にのみページがクリップボードを読むことを許可します。Ctrl+V でテキストをボックスに貼り付けてください。',
        clipboardEmpty: 'クリップボードにテキストがありません。',
        tooLong: 'テキストが長すぎます。上限は {{max}} 文字です。',
        inProgress: '別の貼り付けを入力中です。',
        typing: 'ホストに入力中',
        done: '入力完了',
        canceled: '貼り付けをキャンセルしました',
        failed: '貼り付けに失敗しました',
        cancel: 'キャンセル',
        controlBusy: '別のコントローラーがキーボードを使用中です。',
        hidError: 'キー入力をホストに送信できませんでした。'
      },
      shortcut: {
        title: 'ショートカット',
        custom: 'カスタム',
        capture: 'ショートカットをキャプチャするにはここをクリックしてください',
        clear: 'クリア',
        save: '保存',
        captureTips:
          'Windows キーなどのシステムレベルのキーを取得するには、全画面表示の許可が必要です。',
        enterFullScreen: '全画面モードに切り替えます。'
      },
      leaderKey: {
        title: 'リーダーキー',
        desc: 'ブラウザの制限を回避して、システムによってブロックされているショートカットキーをリモートホストに送信します。',
        howToUse: '使用方法',
        simultaneous: {
          title: '同時モード',
          desc1: 'リーダーキーを押したまま、ショートカットを押します。',
          desc2:
            '操作は直感的ですが、システムの使用状況により一部のショートカットキーが機能しない場合があります。'
        },
        sequential: {
          title: 'シーケンシャルモード',
          desc1: 'リーダーキーを押す → ショートカットを順番に押す → リーダーキーをもう一度押す。',
          desc2: 'いくつかの手順が必要ですが、システムキーの競合を完全に回避します。'
        },
        enable: 'リーダーキーを有効化',
        tip: 'リーダーキーに設定すると、このキーはショートカットのトリガー専用になり、通常の動作は失われます。',
        placeholder: 'リーダーキーを押してください',
        shiftRight: '右 Shift',
        ctrlRight: '右 Ctrl',
        metaRight: '右 Win',
        submit: '送信',
        recorder: {
          rec: 'REC',
          activate: 'キーを有効化',
          input: 'ショートカットキーを押してください...'
        }
      }
    },
    mouse: {
      title: 'マウス',
      cursor: 'ポインター形状',
      default: 'デフォルトポインター',
      pointer: 'ポインターカーソル',
      cell: 'セルポインター',
      text: 'テキストカーソル',
      grab: 'つかむポインター',
      hide: 'ポインターを非表示',
      mode: 'マウスモード',
      absolute: '絶対モード',
      relative: '相対モード',
      absoluteShort: '絶対',
      relativeShort: '相対',
      absoluteStalled: 'ターゲットが絶対マウスを無視しています',
      absoluteStalledDesc:
        'ターゲットが絶対マウスのレポートを受け取らなくなったため、ポインターの移動が失われています。キーボードには影響ありません。USB を復旧すると解消することが多く、相対モードは別のエンドポイントを使用します。',
      useRelative: '相対モードに切り替え',
      direction: 'ホイール方向',
      scrollUp: '上',
      scrollDown: '下',
      speed: 'ホイール速度',
      fast: '速い',
      slow: '遅い',
      requestPointer:
        '相対モードを使用中です。マウスポインターを取得するには、デスクトップをクリックしてください。',
      resetHid: 'HID をリセット',
      hidOnly: {
        title: 'HID-Only モード',
        desc: '使用中にマウスとキーボードが反応しなくなり、HID をリセットしても効果がない場合は、NanoKVM とデバイス間の互換性に問題がある可能性があります。互換性を向上させるために、HID-Only モードを有効にすることをお勧めします。',
        tip1: 'HID-Only モードを有効にすると、仮想 U ディスクと仮想ネットワークがアンマウントされます',
        tip2: 'HID-Only モードでは、イメージのマウントは無効になります',
        rebuild: 'モードを切り替えると USB 接続が再構築されます。NanoKVM は再起動しません',
        enable: 'HID-Only モードを有効化',
        disable: 'HID-Only モードを無効化'
      }
    },
    image: {
      title: 'イメージ',
      loading: '読み込み中',
      empty: 'イメージファイルがありません',
      mountMode: 'マウントモード',
      mountFailed: 'マウントに失敗しました',
      mountDesc:
        '一部のシステムでは、イメージをマウントする前にリモートホストで仮想ディスクをアンマウントする必要があります。',
      unmountFailed: 'アンマウントに失敗しました',
      unmountDesc:
        '一部のシステムでは、イメージをアンマウントする前にリモートホストから手動で取り出す必要があります。',
      refresh: 'イメージリストを更新',
      disk: 'ディスク',
      cdrom: 'CD',
      driveEmpty: '空',
      eject: '取り出し',
      readOnly: '読み取り専用',
      readOnlyTip: '次にディスクに挿入するイメージに適用されます。',
      noDrives: '仮想ドライブがありません。設定で仮想ディスクを有効にしてください。',
      insertFailed: '挿入に失敗しました',
      ejectFailed: '取り出しに失敗しました',
      insertInto: '{{drive}} に挿入します。クリックして変更します。',
      loadedIn: '{{drive}} ドライブに挿入済み',
      attention: '注意',
      deleteConfirm: 'このイメージを削除してもよろしいですか？',
      okBtn: 'はい',
      cancelBtn: 'いいえ',
      deleteFailed: '削除に失敗しました',
      ventoy: {
        statusNoKernel: 'このファームウェアは非対応',
        statusNotInstalled: '未インストール',
        statusReady: '準備完了',
        statusSelected: '選択したイメージ: {{count}}',
        statusInDrive: 'ディスクドライブ内, {{size}}',
        noKernel:
          'このファームウェアのカーネルは device-mapper に対応していないため、対応したイメージをインストールするまで Ventoy は使えません。',
        installDesc: '複数のイメージをコピーせずに 1 つのディスクにまとめ、ホストを起動します。',
        install: 'インストール',
        installing: 'Ventoy をダウンロード中です (約 20 MB)。数分かかることがあります。',
        needsData: 'Ventoy には /data パーティションがマウントされた IronKVM イメージが必要です。',
        uninstall: 'アンインストール',
        uninstallConfirm: 'Ventoy のファイルを削除しますか?',
        noImages: 'Ventoy ディスクに追加できるイメージがありません。',
        onDisk: 'Ventoy ディスクに含める',
        missing: '見つかりません: {{file}}',
        remove: 'Ventoy ディスクから外す',
        setHint:
          'イメージの組み合わせは、Ventoy ディスクがドライブに入っていないときだけ変更できます。',
        useAsDisk: '仮想ディスクとして使用',
        failed: 'Ventoy の要求に失敗しました',
        secureBoot:
          'Secure Boot が有効な場合、ホストで一度だけ MokManager に Ventoy の鍵を登録する必要があります。鍵ファイル ENROLL_THIS_KEY_IN_MOKMANAGER.cer は VTOYEFI パーティションにあります。',
        readOnly:
          'ホストからはディスクが読み取り専用に見えるため、Ventoy の永続化とドライブ上の ventoy.json は機能しません。'
      },
      tips: {
        title: 'アップロード方法',
        usb1: 'NanoKVM を USB 経由でコンピュータに接続します；',
        usb2: '仮想ディスクがマウントされていることを確認します（設定 - 仮想ディスク）；',
        usb3: 'コンピュータ上で仮想ディスクを開き、イメージファイルを仮想ディスクのルートディレクトリにコピーします。',
        scp1: 'NanoKVM とコンピュータが同じローカルエリアネットワークに接続されていることを確認します；',
        scp2: 'コンピュータでターミナルを開き、SCP コマンドを使用してイメージファイルを NanoKVM の /data ディレクトリにアップロードします。',
        scp3: '例：scp your-image-path root@your-nanokvm-ip:/data',
        tfCard: 'TF カード',
        tf1: 'この方法は Linux システムでサポートされています',
        tf2: 'NanoKVM から TF カードを取り出します（フルバージョンでは、まずケースを分解してください）；',
        tf3: 'TF カードをカードリーダーに挿入してコンピュータに接続します；',
        tf4: 'コンピューターから TF カードの /data ディレクトリにイメージファイルをコピーします；',
        tf5: 'TF カードを NanoKVM に挿入します。'
      }
    },
    script: {
      title: 'スクリプト',
      upload: 'アップロード',
      run: '実行',
      runBackground: 'バックグラウンドで実行',
      runFailed: '実行に失敗しました',
      attention: '注意',
      delDesc: 'このファイルを削除してもよろしいですか？',
      confirm: 'はい',
      cancel: 'いいえ',
      delete: '削除',
      close: '閉じる'
    },
    terminal: {
      title: 'ターミナル',
      nanokvm: 'NanoKVM ターミナル',
      serial: 'シリアルポートターミナル',
      serialPort: 'シリアルポート',
      serialPortPlaceholder: 'シリアルポートを入力してください',
      baudrate: 'ボーレート',
      parity: 'パリティ',
      parityNone: 'なし',
      parityEven: '偶数',
      parityOdd: '奇数',
      flowControl: 'フロー制御',
      flowControlNone: 'なし',
      flowControlSoft: 'ソフトウェア',
      flowControlHard: 'ハードウェア',
      dataBits: 'データビット',
      stopBits: 'ストップビット',
      confirm: 'OK'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'コマンドを送信中...',
      sent: 'コマンドを送信しました',
      input: 'MAC アドレスを入力してください',
      ok: 'OK'
    },
    download: {
      title: 'イメージダウンローダー',
      input: 'リモートイメージの URL を入力してください',
      ok: 'OK',
      disabled:
        '/data パーティションは読み取り専用であり、イメージのダウンロードには使用できません',
      uploadbox: 'ここにファイルをドロップするか、クリックして選択してください',
      inputfile: '画像ファイルを入力してください',
      NoISO: 'ISO なし',
      sha256: 'SHA-256（任意）',
      sha256Placeholder: '64 文字の SHA-256 チェックサムを入力してください',
      invalidSHA256: 'SHA-256 は 64 文字の 16 進数文字列である必要があります',
      failed: 'ダウンロードに失敗しました',
      success: 'ダウンロードに成功しました',
      checksumFailed: 'ダウンロードに失敗しました：SHA-256 検証に失敗しました',
      cancel: 'キャンセル',
      cancelFailed: 'ダウンロードのキャンセルに失敗しました',
      bootMenu: 'ブートメニュー (netboot.xyz)',
      bootMenuDesc: '仮想 CD 用に netboot.xyz の ISO をチェックサム検証付きでダウンロード'
    },
    power: {
      title: '電源',
      showConfirm: '確認メッセージ',
      showConfirmTip: '電源操作の確認メッセージを表示する',
      reset: 'リセット',
      power: '電源',
      powerShort: '電源（クリック）',
      powerLong: '電源（長押し）',
      resetConfirm: '再起動を実行しますか？',
      powerConfirm: '電源操作を実行しますか？',
      okBtn: 'はい',
      cancelBtn: 'いいえ',
      hostOs: 'ホスト OS',
      hostOsTip: 'USB キーとして送信されます。動作はホストが決めます。',
      sleep: 'スリープ',
      wake: 'スリープ解除',
      wakeKey: 'Shift でスリープ解除',
      powerDown: 'シャットダウン',
      sleepConfirm: 'ホストをスリープさせますか？',
      powerDownConfirm: '電源オフキーをホストに送信しますか？',
      wakeTip:
        'スリープ中のホストは、自分をスリープさせたデバイスからのスリープ解除を無視することがよくあります。「Shift でスリープ解除」はキーボードのキーを押すため、より多くのホストが応答します。',
      led: '電源 LED',
      ledOn: '点灯',
      ledOff: '消灯',
      ledUnknown: '不明',
      ledConnected: '電源 LED 接続済み',
      ledConnectedTip:
        'ホストの電源 LED ヘッダーがボードに配線されている場合のみオンにしてください。配線がないと電源状態は不明になります。',
      ledConnectedFailed: '電源 LED 設定の保存に失敗しました'
    },
    settings: {
      title: '設定',
      mcp: {
        title: 'MCP サービス',
        service: 'MCP リモート制御',
        serviceDesc:
          '信頼できる MCP クライアントによるキーボードとマウスの操作、およびスクリーンショットの取得を許可します',
        securityWarning:
          'この API キーを持つ人は誰でもリモートホストを操作し、画面を表示できます。HTTPS を使用し、信頼できるネットワークでのみ有効にしてください。',
        endpoint: 'エンドポイント',
        apiKey: 'API キー',
        regenerateConfirmTitle: 'MCP API キーを再生成しますか？',
        regenerateConfirmDesc: '現在のキーは直ちに使用できなくなります。',
        enableConfirmTitle: '外部 MCP 制御を有効にしますか？',
        enableConfirmDesc:
          'MCP を有効にすると PicoClaw が停止し、アクティブな PicoClaw セッションがすべて終了します。',
        failed: 'MCP 操作に失敗しました',
        copyFailed: 'コピーに失敗しました。手動でコピーしてください。',
        okBtn: '確認',
        cancelBtn: 'キャンセル'
      },
      redfish: {
        title: 'Redfish',
        service: 'Redfish サービス',
        serviceDesc:
          'DMTF Redfish API です。redfishtool や Ansible などのツールから電源制御、仮想メディア、ステータスを利用できます。オフにすると、すべての Redfish セッションが終了します。',
        endpoint: 'サービスルート',
        httpsOn: 'ボードは HTTPS で提供しています。多くの Redfish ツールでは HTTPS が必要です。',
        httpsOff:
          'ボードは通常の HTTP で提供しています。多くの Redfish ツールでは HTTPS が必要です。「設定 > ネットワーク」で有効にしてください。',
        credentials:
          'Redfish では KVM のアカウント（Basic 認証または Redfish セッション）と、X-Auth-Token として送信する API キーを使用できます。API キーは API キーページで管理します。',
        powerActions: '電源操作',
        powerActionsDesc:
          '現在提供されているリセットタイプです。On、ForceOff、GracefulShutdown は電源状態が必要なため、電源メニューで「電源 LED 接続済み」がオンの場合にのみ提供されます。',
        sessions: 'セッション',
        noSessions: '開いている Redfish セッションはありません',
        created: '作成日時',
        lastUsed: '最終使用',
        refresh: '更新',
        end: '終了',
        endConfirmTitle: 'この Redfish セッションを終了しますか？',
        endConfirmDesc:
          'このセッションのトークンは直ちに使用できなくなります。クライアントは再度ログインする必要があります。',
        failed: 'Redfish の操作に失敗しました',
        copyFailed: 'コピーに失敗しました。手動でコピーしてください。',
        okBtn: '確認',
        cancelBtn: 'キャンセル'
      },
      watchdog: {
        title: 'ウォッチドッグ',
        service: 'ホストウォッチドッグ',
        serviceDesc:
          'ホストが起動しているはずなのに、タイムアウトの間に画面が変化しない、または HDMI 信号がない場合、ボードがリセットを押すか、ホストの電源を入れ直します。',
        stillWarning:
          'ディスプレイがスリープするホストや、動作中も画面が静止しているホストは、ハングしているように見えます。ホストのディスプレイスリープをオフにするか、Ping アドレスを設定してください。',
        ledHint:
          '電源メニューの「電源 LED 接続済み」がオフです。ウォッチドッグはホストの電源が切れていることを検出できないため、ホストを常にオンとして扱います。',
        timeout: 'タイムアウト',
        timeoutDesc: 'ウォッチドッグが動作するまで、ホストが生存の兆候を示さなくてもよい時間。',
        action: '動作',
        actionDesc: '電源の入れ直しは電源ボタンを 5 秒間押し続け、その後もう一度押します。',
        actionReset: 'リセット',
        actionPower: '電源の入れ直し',
        cooldown: 'クールダウン',
        cooldownDesc: '2 回の動作の最短間隔。',
        maxPerHour: '1 時間あたりの動作回数',
        maxPerHourDesc: '1 時間あたりの動作の最大回数。',
        pingHost: 'Ping アドレス',
        pingHostDesc:
          'ホストの IP アドレス。応答は生存の兆候として扱われます。Ping しない場合は空欄にしてください。',
        pingHostInvalid: 'IPv4 または IPv6 アドレスを入力してください。',
        minutes: '分',
        save: '保存',
        saved: '保存しました',
        state: '検出器',
        status: {
          off: 'オフ',
          watching: '監視中',
          hostOff: 'ホスト電源オフ',
          captureOff: 'HDMI キャプチャオフ',
          cooldown: 'クールダウン中',
          capped: '1 時間の上限に到達',
          acting: '動作中'
        },
        signal: 'HDMI 信号',
        yes: 'あり',
        no: 'なし',
        led: '電源 LED',
        on: '点灯',
        off: '消灯',
        ledNotConnected: '未接続',
        ping: 'Ping',
        pingNotSet: '未設定',
        pingReply: '応答あり',
        pingNoReply: '応答なし',
        lastChange: '最後の画面変化',
        never: 'なし',
        actsIn: '動作まで',
        actionsLastHour: '直近 1 時間の動作',
        duration: '{{minutes}} 分 {{seconds}} 秒',
        log: 'ログ',
        noLog: 'ウォッチドッグはまだ動作していません。',
        refresh: '更新',
        reasonFrozen: '画面が変化しなかった',
        reasonNoSignal: 'HDMI 信号なし',
        stuckFor: '{{duration}} 生存の兆候なし',
        pressFailed: 'ボタン操作に失敗しました: {{error}}',
        noScreenshot: 'スクリーンショットなし',
        failed: 'ウォッチドッグの操作に失敗しました'
      },
      netboot: {
        title: 'ネットワークブート',
        description:
          'ホストをネットワークから起動します。USB ネットワークリンク経由で iPXE と KVM 上のイメージのメニュー、または LAN 上のプロキシ DHCP で netboot.xyz を提供します。',
        addon: 'dnsmasq とブートファイル',
        addonDesc:
          '/data にインストールされます。dnsmasq は Alpine から、iPXE と netboot.xyz は各リリースから取得し、それぞれチェックサムで検証します。',
        install: 'インストール',
        installing: 'インストール中です。数分かかることがあります。',
        uninstall: 'アンインストール',
        uninstallConfirm:
          'ネットワークブートをオフにして、dnsmasq とブートファイルを削除しますか？',
        needsData:
          'ネットワークブートには /data パーティションがマウントされた IronKVM イメージが必要です。',
        usb: 'USB ネットワークリンク上',
        usbDesc:
          'USB ネットワークリンクがオンの間、udhcpd の代わりに dnsmasq が応答します。ホストはルーターも DNS サーバーもない 1 つのアドレス、アーキテクチャに合った iPXE、KVM 上の ISO イメージのメニューを受け取ります。',
        linkOff:
          'USB ネットワークリンクがオフです。デバイスの USB ネットワークでオンにしてください。',
        menuUrl: 'メニュー',
        leases: 'ホストのリース',
        noLeases: 'まだありません',
        netbootxyzNote:
          'メニューの netboot.xyz はインターネットから読み込まれますが、USB リンクはインターネットに届きません。ホストの別のネットワークポートにインターネット接続が必要です。',
        lan: 'LAN 上のプロキシ DHCP',
        lanDesc:
          'LAN 上の PXE クライアントに netboot.xyz を提供し、netboot.xyz はメニューをインターネットから読み込みます。アドレスは配布せず、KVM 上のイメージも提供しません。',
        lanWarning:
          'ホストだけでなく、この LAN 上のすべての PXE クライアントに netboot.xyz が提供されます。自分で管理しているネットワークでのみオンにしてください。',
        lanConfirm: 'LAN 上のプロキシ DHCP をオンにしますか？',
        lanInterface: 'LAN',
        running: '実行中',
        stopped: '停止中',
        images: 'メニュー内のイメージ',
        noImages: 'イメージディレクトリに ISO イメージがありません。',
        boots: '最近のブート',
        noBoots: 'ホストはまだ何も取得していません。',
        log: 'dnsmasq ログ',
        refresh: '更新',
        okBtn: '確認',
        cancelBtn: 'キャンセル',
        failed: 'ネットワークブートの操作に失敗しました'
      },
      about: {
        title: 'NanoKVM について',
        information: '情報',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'アプリケーションバージョン',
        applicationTip: 'NanoKVM ウェブアプリケーションバージョン',
        image: 'イメージバージョン',
        imageTip: 'NanoKVM システムイメージバージョン',
        kernel: 'カーネルバージョン',
        kernelTip: '現在実行中の Linux カーネルのリリース',
        deviceKey: 'デバイスキー',
        videoMemory: 'ビデオメモリ',
        videoMemoryTip:
          'ビデオキャプチャ用に予約されたメモリです。システムの他の部分とは共有されません。',
        videoMemoryGenerations_other:
          '以前の NanoKVM セッション {{count}} 件がビデオメモリを保持しています',
        videoMemoryReboot: '再起動すると回収されます。',
        community: 'コミュニティ',
        hostname: 'ホスト名',
        hostnameUpdated: 'ホスト名は正常に変更され、再起動後に有効になります',
        ipType: {
          Wired: '有線',
          Wireless: 'ワイヤレス',
          Other: 'その他'
        }
      },
      appearance: {
        title: '外観',
        display: '表示',
        language: '言語',
        languageDesc: 'インターフェース言語の選択',
        webTitle: 'ウェブページタイトル',
        webTitleDesc: 'ウェブページタイトルのカスタマイズ',
        menuBar: {
          title: 'メニューバー',
          mode: '表示モード',
          modeDesc: 'メニューバーの画面表示方法',
          modeOff: '閉じる',
          modeAuto: '自動非表示',
          modeAlways: '常に表示',
          keyboardLedStatus: 'キーボードロックの表示',
          keyboardLedStatusDesc:
            'リモートコンピューターの Num Lock、Caps Lock、Scroll Lock の状態を表示',
          icons: 'メニューアイコン',
          iconsDesc: 'メニューバーでのサブメニューアイコンの表示'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'リモートキーボードのロック状態',
        indicatorLabel: '{{label}}：{{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'オン',
        off: 'オフ',
        unknown: '不明'
      },
      device: {
        title: 'デバイス',
        oled: {
          title: 'OLED',
          description: 'OLED 画面の自動スリープ時間',
          brightness: 'OLED の明るさ',
          brightnessDescription: '明るさを下げるとパネルが長持ちします',
          brightnessLevels: {
            '64': '最低',
            '96': '低',
            '128': '中',
            '160': '高',
            '207': 'デフォルト',
            '255': '最大'
          },
          0: '無効',
          15: '15秒',
          30: '30秒',
          60: '1分',
          180: '3分',
          300: '5分',
          600: '10分',
          1800: '30分',
          3600: '1時間'
        },
        ssh: {
          description: 'SSH リモートアクセスを有効にする',
          tip: '使用する前に必ず強力なパスワードを設定してください（アカウント - パスワードの変更）'
        },
        advanced: '詳細設定',
        cpuFreq: {
          title: 'CPU 周波数',
          description: '次回起動時に適用する CPU クロックを設定する',
          tip: 'CPU は 850 MHz で起動し、定格は 1000 MHz です。新しい値はシステムの動作中ではなく、次回起動時に適用されます。1000 MHz は仕様の範囲内で、どちらの設定でも温度は上限を十分に下回ります。',
          running: '動作中: {{mhz}} MHz',
          rebootToApply: '再起動して適用',
          rebootConfirm: '今すぐ再起動して {{mhz}} MHz を適用しますか？'
        },
        swap: {
          title: 'スワップ',
          disable: '無効',
          description: 'スワップファイルのサイズを設定する',
          tip: 'この機能を有効にすると、SD カードの寿命が短くなる可能性があります！'
        },
        zram: {
          title: '圧縮スワップ（zram）',
          description: 'SD カードではなく、圧縮した RAM 上でスワップする',
          tip: 'zram はスワップを SD カードに置かないため、カードを消耗させません。背後にディスクのスワップはないため、zram がいっぱいになると、カーネルはゆっくりページングする代わりにプロセスを停止します。メモリ上限は zram が使用できる RAM の量を制限します。',
          unavailable: 'このデバイスにはカーネルモジュールがインストールされていません',
          inactive: '有効ですが、デバイスが起動しませんでした',
          active: '動作中 - {{used}} / {{total}}、{{ratio}}x',
          off: 'オフ',
          detail: {
            algorithm: 'アルゴリズム: {{algorithm}}',
            memory: '使用メモリ: {{used}} / {{limit}}',
            memoryNoLimit: '使用メモリ: {{used}}、上限なし',
            counters:
              'スワップイン {{in}} ページ、スワップアウト {{out}} ページ（全スワップデバイス、起動以降）'
          }
        },
        mouseJiggler: {
          title: 'マウスジグラー',
          description: 'リモートホストの休止を防ぐ',
          disable: '閉じる',
          absolute: '絶対モード',
          relative: '相対モード'
        },
        mdns: {
          description: 'mDNS 検出サービスを有効にする',
          tip: 'この機能を使用していない場合は、オフにすることをお勧めします'
        },
        hdmi: {
          description: 'HDMI/モニター 出力機能を有効にする',
          idleTimeoutTitle: 'キャプチャのアイドルタイムアウト',
          idleTimeoutDescription:
            'アクティブな閲覧者がいない状態が次の時間続いたら HDMI キャプチャを停止',
          minutes: '分'
        },
        autostart: {
          title: '自動起動スクリプト設定',
          description: 'NanoKVM の起動時に自動的に実行されるスクリプトファイルを管理します',
          new: '新しいスクリプトを作成する',
          deleteConfirm: 'このファイルを削除してもよろしいですか？',
          yes: 'はい',
          no: 'いいえ',
          scriptName: '自動起動スクリプト名',
          scriptContent: '自動起動スクリプト内容',
          settings: '設定'
        },
        hidOnly: 'HID-Only モード',
        hidOnlyDesc:
          'このモードでは仮想デバイスはマウントされなくなり、基本的な HID 制御機能のみが保持されます。',
        disk: '仮想ディスク',
        diskDesc: 'リモートホストに仮想 USB ドライブをマウントする',
        network: '仮想ネットワークカード',
        networkDesc: 'リモートホストに仮想ネットワークカードをマウントする',
        usbNetwork: {
          description:
            'USB ケーブル経由でリモートホストと結ぶプライベートネットワークです。ホストにはゲートウェイと DNS のないアドレスが割り当てられるため、NanoKVM を経由して LAN に到達することはできません。',
          off: 'オフ',
          ncm: 'NCM (Linux、macOS、Windows 11)',
          ecm: 'ECM (NCM 非対応のホスト向け)',
          rndis: 'RNDIS (提供終了)',
          rndisNote:
            'この接続は提供を終了した RNDIS を使用しています。NCM または ECM を選択してください。',
          subnet: 'サブネット',
          subnetDesc:
            '/24 から /30 のプライベート IPv4 ネットワーク。NanoKVM が最初のアドレスを、ホストが 2 番目のアドレスを使用します。',
          addresses: 'NanoKVM: {{board}}、ホスト: {{host}}',
          invalidSubnet: '172.31.255.0/30 のようなサブネットを入力してください。',
          apply: '適用',
          confirm: 'USB デバイスを再接続しますか?',
          reenumerate:
            '適用すると USB 接続が再構築されます。ホストは数秒間、キーボード、マウス、仮想ディスクを使用できなくなります。'
        },
        audio: '仮想スピーカー',
        audioDesc:
          'リモートホストに USB サウンドカードを提供し、ホストの音声を聞けるようにします。ホスト側で出力デバイスとして選択する必要があります。切り替えると USB 接続が再構築されます。',
        audioNote:
          '音声は H.264 の両モード（WebRTC と Direct）で利用でき、MJPEG では利用できません',
        console: 'シリアルコンソール',
        consoleDesc:
          'リモートホストに USB シリアルポートを提供し、ネットワークに接続できないときにこの NanoKVM にログインできるようにします',
        consoleTip:
          'リモートホストを操作できる人は誰でも、この NanoKVM のログインプロンプトにアクセスできます。有効にする前に強力なパスワードを設定してください（アカウント - パスワードの変更）。',
        endpoints: {
          title: 'USB エンドポイント',
          used: '{{total}} 個中 {{used}} 個使用',
          cost: '{{cost}} 個使用',
          needs: '{{cost}} 個必要',
          full: 'USB エンドポイントが不足しています。先に他の機能をオフにしてください。',
          inactive:
            'オンですが動作していません。USB コントローラーのエンドポイントが不足しています。他のデバイスをオフにすると、すぐに起動します。',
          explain:
            'USB コントローラーの入力エンドポイントの数は固定されており、ここではその数を数えています。収まる数より多くのデバイスが有効な場合、キーボードとマウスは維持され、残りはオフになります。',
          error: 'デバイスに接続できませんでした。もう一度お試しください。',
          fitTogether: '同時に使用できる組み合わせ: {{sets}}'
        },
        reboot: '再起動',
        rebootDesc: 'NanoKVM を再起動してもよろしいですか?',
        okBtn: 'はい',
        cancelBtn: 'いいえ'
      },
      network: {
        title: 'ネットワーク',
        wifi: {
          title: 'Wi-Fi',
          description: 'Wi-Fi 設定',
          apMode: 'AP モードが有効になりました。QR コードをスキャンして Wi-Fi に接続してください。',
          connect: 'Wi-Fi に接続',
          connectDesc1: 'SSID とパスワードを入力してください',
          connectDesc2: 'このネットワークに接続するためのパスワードを入力してください',
          disconnect: 'このネットワーク接続を切断しますか？',
          failed: '接続に失敗しました。もう一度お試しください。',
          ssid: 'SSID',
          password: 'パスワード',
          joinBtn: '接続',
          confirmBtn: 'OK',
          cancelBtn: 'キャンセル'
        },
        tls: {
          description: 'HTTPS プロトコルを有効にする',
          tip: '注意：HTTPS を使用すると、特に MJPEG ビデオモードで遅延が増加する可能性があります。',
          restarting: 'デバイスのサーバーを再起動しています。約 2 分かかります...',
          waiting: 'デバイスの応答を待っています...',
          waitingHttp:
            'http に戻しています。自動的に開かない場合は、このページを再読み込みしてください。'
        },
        ethernet: {
          title: 'IPアドレス',
          description: 'NanoKVM が有線ネットワークでアドレスを取得する方法を設定します',
          dhcp: 'DHCP',
          manual: '手動',
          networkDetails: 'ネットワークの詳細',
          interface: 'インターフェース',
          ipAddress: 'IPアドレス',
          subnetMask: 'サブネットマスク',
          router: 'ルーター',
          save: '適用',
          invalidAddress: '有効なIPアドレスを入力してください',
          invalidMask: '有効なサブネットマスクを入力してください。例: 255.255.255.0 または 24',
          invalidRouter: '有効なルーターアドレスを入力してください',
          addressRequired: 'IPアドレスが必要です',
          maskRequired: 'サブネットマスクが必要です',
          applyTitle: 'NanoKVM のアドレスを変更しますか?',
          applyWarning:
            'このページとの接続は失われます。NanoKVM は新しいアドレスを適用し、そこに到達するのを {{seconds}} 秒間待ちます。到達すれば変更が保持されます。何も到達しない場合、NanoKVM は以前の設定に戻します。',
          applyConfirm: '適用',
          applyCancel: 'キャンセル',
          applyFailed: 'アドレスを適用できませんでした',
          trialTitle: '確認を待っています',
          trialDhcp: 'NanoKVM は DHCP にアドレスを要求しています。',
          trialStatic: 'NanoKVM は現在 {{address}} にあります。',
          trialInstruction:
            '新しいアドレスで NanoKVM を開き、求められたらサインインしてください。そこに到達すれば変更が保持されます。{{seconds}} 秒以内に何も到達しない場合、以前の設定に戻ります。',
          trialOpen: '新しいアドレスを開く',
          trialKeep: 'この設定を保持',
          trialKept: '新しいアドレスを保存しました',
          trialKeepFailed: '設定を保持できませんでした',
          trialGone: '変更はすでに元に戻されています。もう一度お試しください。',
          unsaved: '未保存の変更'
        },
        dns: {
          title: 'DNS',
          description: 'NanoKVM の DNS サーバーを設定',
          mode: 'モード',
          dhcp: 'DHCP',
          manual: '手動',
          add: 'DNS を追加',
          save: '保存',
          invalid: '有効な IP アドレスを入力してください',
          noDhcp: '現在 DHCP DNS は利用できません',
          saved: 'DNS 設定を保存しました',
          saveFailed: 'DNS 設定の保存に失敗しました',
          unsaved: '未保存の変更',
          maxServers: 'DNS サーバーは最大 {{count}} 個までです',
          dnsServers: 'DNS サーバー',
          dhcpServersDescription: 'DNS サーバーは DHCP から自動取得されます',
          manualServersDescription: 'DNS サーバーは手動で編集できます',
          networkDetails: 'ネットワーク詳細',
          interface: 'インターフェイス',
          ipAddress: 'IP アドレス',
          subnetMask: 'サブネットマスク',
          router: 'ルーター',
          none: 'なし'
        }
      },
      vpn: {
        loading: '読み込み中...',
        okBtn: 'はい',
        cancelBtn: 'いいえ',
        restart: '{{name}} を再起動しますか？',
        stop: '{{name}} を停止しますか？',
        stopDesc:
          'デーモンは今すぐ停止します。起動時に開始は別のスイッチで、現在の設定のまま変わりません。',
        update: '{{name}} を {{version}} にアップデートしますか？',
        updateDesc: 'デーモンが実行中の場合は再起動します。ログイン状態は保持されます。',
        notInstall: '{{name}} がインストールされていません。',
        install: 'インストール',
        installing: 'インストール中',
        installFailed: 'インストールに失敗しました',
        retry: '再試行',
        notRunning: '{{name}} が実行されていません。続行するには開始してください。',
        run: '開始',
        boot: '起動時に開始',
        bootDesc: 'KVM の起動時に {{name}} を開始します。',
        enable: '{{name}} を有効化',
        control: 'コントロールサーバー',
        connected: '接続済み',
        disconnected: '未接続',
        deviceName: 'デバイス名',
        deviceIP: 'デバイス IP',
        account: 'アカウント',
        version: 'バージョン',
        uptime: '稼働時間',
        peers: 'ピア',
        noPeers: 'ピアはまだありません。',
        online: 'オンライン',
        offline: 'オフライン',
        memory: 'メモリ',
        daemonRss: 'デーモン',
        group: 'アドオングループ',
        high: '{{size}} を超えると制限',
        max: '{{size}} を超えるとカーネルが停止',
        noGroup: 'このボードにはアドオン用のメモリグループがありません。',
        uninstall: '{{name}} をアンインストール',
        uninstallDesc:
          '{{name}} をアンインストールしてもよろしいですか？ログイン情報はボードに残ります。',
        blocked:
          '{{other}} が実行中か、起動時に開始する設定になっています。同時に実行できる VPN は 1 つだけです。先に {{other}} を停止し、起動時の開始をオフにしてください。',
        swap: {
          title: 'スワップメモリ',
          tip: 'デーモンのメモリが不足する場合は、スワップメモリを有効にしてみてください。スワップファイルのサイズはデフォルトで 256MB に設定され、「設定 > デバイス」で調整できます。'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'ページを更新してもう一度お試しいただくか、手動でインストールしてください',
        download: 'ダウンロードして',
        package: 'インストールパッケージを',
        unzip: '解凍してください',
        upTailscale: 'tailscale ファイルを NanoKVM の /usr/bin ディレクトリにアップロードします',
        upTailscaled: 'tailscaled ファイルを NanoKVM の /usr/sbin ディレクトリにアップロードします',
        refresh: 'ページを更新します',
        notLogin:
          'このデバイスはまだバインドされていません。ログインしてデバイスをアカウントにバインドしてください。',
        urlPeriod: 'この URL は 10 分間有効です',
        login: 'ログイン',
        loginSuccess: 'ログイン成功',
        logout: 'ログアウト',
        logoutDesc: 'ログアウトしてもよろしいですか？'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'このデバイスはまだ NetBird ネットワークに参加していません。セットアップキーで参加するか、SSO でログインしてください。',
        setupKey: 'セットアップキー',
        setupKeyPlaceholder: 'NetBird ダッシュボードのセットアップキーを貼り付けてください',
        join: '参加',
        or: 'または',
        sso: 'SSO でログイン',
        urlPeriod: 'この URL は 10 分間有効です',
        loginSuccess: 'ログイン成功',
        logout: '登録解除',
        logoutDesc:
          '登録を解除すると、このピアが NetBird アカウントから削除され、ここにある設定も削除されます。再度参加するにはセットアップキーまたは SSO ログインが必要で、ピアに新しい IP が割り当てられる場合があります。続行しますか？'
      },
      update: {
        title: 'アップデート',
        queryFailed: 'バージョン番号の取得に失敗しました',
        updateFailed: 'アップデートに失敗しました。もう一度お試しください。',
        isLatest: 'すでに最新バージョンです。',
        available: '新しいバージョンが利用可能です。アップデートしてもよろしいですか？',
        updating: 'アップデート中、お待ちください...',
        confirm: 'はい',
        cancel: 'いいえ',
        preview: 'プレビューアップデート',
        previewDesc: '新機能や改善をいち早く体験する',
        previewTip:
          'プレビューアップデートには不安定な部分や不完全な機能が含まれる場合があります！',
        customServer: {
          title: 'カスタム更新サーバー',
          desc: '指定したサーバーでオンラインアップデートを確認し、ダウンロードします',
          invalidUrl:
            'クエリ、フラグメント、latest.json を含まない、有効な HTTP または HTTPS のサーバーディレクトリを入力してください。',
          loadFailed: '更新サーバーの設定を読み込めませんでした。',
          saveFailed: '更新サーバーの設定を保存できませんでした。',
          saved: '更新サーバーの設定を保存しました。',
          save: '保存',
          confirmTitle: 'カスタム更新サーバーを使用しますか？',
          confirmDesc:
            'SHA-512 で確認できるのは、パッケージがこのサーバーから提供されたマニフェストと一致することだけです。そのパッケージが NanoKVM の公式リリースであることは保証されません。不具合のあるサーバーや悪意のあるサーバーを使用すると、デバイスが使用不能になったり、データが失われたり、システムが侵害されたりする可能性があります。',
          confirm: 'そのまま使用',
          useSipeed: 'Sipeed 公式サーバーを使用',
          previewDisabled:
            'カスタム更新サーバーが有効な間は、プレビュー版アップデートを利用できません。'
        },
        offline: {
          title: 'オフラインアップデート',
          desc: 'ローカルインストールパッケージでアップデートする',
          upload: 'アップロード',
          checksumPlaceholder: 'SHA-256チェックサム（任意）',
          invalidChecksum: 'SHA-256チェックサムは64文字の16進数である必要があります。',
          checksumMismatch:
            'SHA-256の検証に失敗しました。パッケージが破損している可能性があります。',
          invalidName:
            'ファイル名の形式が正しくありません。GitHub リリースページにアクセスしてインストールパッケージをダウンロードしてください。',
          updateFailed: 'アップデートに失敗しました。もう一度お試しください。'
        }
      },
      account: {
        title: 'アカウント',
        webAccount: 'ウェブアカウント名',
        role: 'ロール',
        roles: { admin: '管理者', user: 'ユーザー' },
        password: 'パスワード',
        updateBtn: '変更',
        logoutBtn: 'ログアウト',
        logoutDesc: 'ログアウトしてもよろしいですか？',
        okBtn: 'はい',
        cancelBtn: 'いいえ',
        users: {
          title: 'ユーザー',
          create: 'ユーザーを作成',
          enabled: '有効',
          disabled: '無効',
          deviceOwner: 'デバイス所有者',
          resetPassword: 'パスワードをリセット',
          delete: '削除',
          deleteConfirm: 'このユーザーを削除し、すべてのセッションを無効にしますか？',
          created: 'ユーザーを作成しました',
          deleted: 'ユーザーを削除しました',
          passwordUpdated: 'パスワードを更新しました',
          loadFailed: 'ユーザーの読み込みに失敗しました',
          saveFailed: 'ユーザーの保存に失敗しました',
          deleteFailed: 'ユーザーの削除に失敗しました'
        }
      },
      apiKeys: {
        title: 'API キー',
        description:
          'キーは所有者として、そのユーザーのロールで動作します。メトリクスと API には Authorization: Bearer <key> として、Redfish には X-Auth-Token として送信してください。',
        name: '名前',
        namePlaceholder: 'キーの用途（例: prometheus）',
        nameRequired: 'キーに名前を付けてください',
        nameTooLong: '名前は 64 文字以内にしてください',
        unnamed: '（名前なし）',
        create: 'キーを作成',
        created: '作成日時',
        owner: '所有者',
        empty: 'API キーはありません',
        newKeyTitle: '新しい API キー',
        newKeyWarning:
          '今すぐキーをコピーしてください。キーは保存されず、再表示できません。紛失した場合は、失効させて新しいキーを作成してください。',
        copy: 'コピー',
        copied: 'コピーしました',
        copyFailed: 'コピーに失敗しました。手動でコピーしてください。',
        done: '完了',
        revoke: '失効',
        revokeConfirmTitle: 'この API キーを失効させますか？',
        revokeConfirmDesc: '「{{name}}」を使用しているものはすべて直ちに動作しなくなります。',
        revoked: 'API キーを失効させました',
        loadFailed: 'API キーの読み込みに失敗しました',
        createFailed: 'API キーの作成に失敗しました',
        revokeFailed: 'API キーの失効に失敗しました',
        cancelBtn: 'キャンセル'
      }
    },
    picoclaw: {
      title: 'PicoClaw アシスタント',
      empty: 'パネルを開いてタスクを開始して開始します。',
      inputPlaceholder: 'PicoClaw に実行してほしいことを説明してください',
      newConversation: '新しい会話',
      processing: '処理中...',
      agent: {
        defaultTitle: '一般アシスタント',
        defaultDescription: '一般的なチャット、検索、およびワークスペースのヘルプ。',
        kvmTitle: 'リモート操作',
        kvmDescription: 'NanoKVM を通じてリモート ホストを操作します。',
        switched: 'エージェントの役割が切り替わりました',
        switchFailed: 'エージェントの役割を切り替えることができませんでした'
      },
      send: '送信',
      cancel: 'キャンセル',
      status: {
        connecting: 'ゲートウェイに接続しています...',
        connected: 'PicoClaw セッションが接続されました',
        disconnected: 'PicoClaw セッションが終了しました',
        stopped: '停止要求が送信されました',
        runtimeStarted: 'PicoClaw ランタイムが開始されました',
        runtimeStartFailed: 'PicoClaw ランタイムの開始に失敗しました',
        runtimeStopped: 'PicoClaw ランタイムが停止しました',
        runtimeStopFailed: 'PicoClaw ランタイムの停止に失敗しました',
        controlSwitchedToMCP: '制御が外部 MCP サービスに切り替わりました'
      },
      connection: {
        runtime: {
          checking: 'チェック中',
          restoring: 'Restoring PicoClaw',
          ready: 'ランタイムの準備が完了しました',
          stopped: 'ランタイムが停止しました',
          blockedByMCP: '外部 MCP 制御が有効です',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'ランタイムが使用できません',
          configError: '構成エラー'
        },
        transport: {
          connecting: '接続中',
          connected: '接続されました',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'アイドル状態',
          busy: '忙しいです'
        }
      },
      message: {
        toolAction: 'アクション',
        observation: '観察',
        screenshot: 'スクリーンショット'
      },
      overlay: {
        locked: 'PicoClaw がデバイスを制御しています。手動入力が一時停止されます。'
      },
      control: {
        picoclaw: 'デバイス制御: PicoClaw',
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'デバイス制御: 外部 MCP',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'デバイス制御: オフ',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: '制御を付与',
        release: '解除',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'PicoClaw 制御を付与しました',
        released: 'PicoClaw 制御を解除しました',
        grantFailed: 'PicoClaw 制御の付与に失敗しました',
        releaseFailed: 'PicoClaw 制御の解除に失敗しました',
        grantConfirmTitle: 'デバイス制御を PicoClaw に切り替えますか?',
        grantConfirmDesc: '外部 MCP のデバイス書き込みは中断されます。'
      },
      install: {
        install: 'PicoClaw をインストールする',
        installing: 'PicoClaw をインストールしています',
        success: 'PicoClaw は正常にインストールされました',
        failed: 'PicoClaw のインストールに失敗しました',
        uninstalling: 'ランタイムをアンインストールしています...',
        uninstalled: 'ランタイムは正常にアンインストールされました。',
        uninstallFailed: 'アンインストールに失敗しました。',
        requiredTitle: 'PicoClaw がインストールされていません',
        requiredDescription:
          'PicoClaw ランタイムを開始する前に PicoClaw をインストールしてください。',
        progressDescription: 'PicoClaw をダウンロードしてインストールしています。',
        stages: {
          preparing: '準備中',
          downloading: 'ダウンロード中',
          extracting: '展開中',
          verifying: '検証中',
          installing: 'インストール中',
          installed: 'インストール完了',
          install_timeout: 'タイムアウト',
          install_failed: '失敗'
        }
      },
      model: {
        requiredTitle: 'モデル構成が必要です',
        requiredDescription: 'PicoClaw チャットを使用する前に、PicoClaw モデルを構成します。',
        docsTitle: '構成ガイド',
        docsDesc: 'サポートされているモデルとプロトコル',
        menuLabel: 'モデルの構成',
        modelIdentifier: 'モデル識別子',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API キー',
        apiKeyPlaceholder: 'モデルの API キーを入力してください',
        save: '保存',
        saving: '保存中',
        saved: 'モデル構成が保存されました',
        saveFailed: 'モデル構成の保存に失敗しました',
        invalid: 'モデル識別子、API Base URL、API キーは必須です'
      },
      uninstall: {
        menuLabel: 'アンインストール',
        confirmTitle: 'PicoClaw のアンインストール',
        confirmContent:
          'PicoClaw をアンインストールしてもよろしいですか?これにより、実行可能ファイルとすべての構成ファイルが削除されます。',
        confirmOk: 'アンインストール',
        confirmCancel: 'キャンセル'
      },
      history: {
        title: '履歴',
        loading: 'セッションを読み込み中...',
        emptyTitle: '履歴はまだありません',
        emptyDescription: '以前の PicoClaw セッションがここに表示されます。',
        loadFailed: 'セッション履歴のロードに失敗しました',
        deleteFailed: 'セッションの削除に失敗しました',
        deleteConfirmTitle: 'セッションを削除します',
        deleteConfirmContent: '「{{title}}」を削除してもよろしいですか?',
        deleteConfirmOk: '削除',
        deleteConfirmCancel: 'キャンセル',
        messageCount_other: '{{count}} メッセージ',
        messageCount: '{{count}} メッセージ'
      },
      config: {
        startRuntime: 'PicoClaw を開始',
        stopRuntime: 'PicoClaw を停止'
      },
      start: {
        enableConfirmTitle: '制御を PicoClaw に切り替えますか？',
        enableConfirmDesc: 'PicoClaw を開始すると外部 MCP サービスが無効になります。',
        enableConfirmOk: 'PicoClaw を開始',
        enableConfirmCancel: 'キャンセル',
        title: 'PicoClaw を開始',
        description: 'ランタイムを起動して、PicoClaw アシスタントの使用を開始します。',
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'エラーが発生しました',
      refresh: '更新',
      panel: 'ページのこの部分が動作しなくなりました',
      retry: '再試行'
    },
    fullscreen: {
      toggle: '全画面表示切り替え'
    },
    input: {
      disconnected: 'キーボードとマウスが接続されていません',
      disconnectedTls:
        'キーボードとマウスを送る安全な接続を、ブラウザーが確認なしに拒否しました。このデバイスが生成した証明書がまだ信頼されていません。このアドレスを新しいタブで開いて証明書を受け入れ、再読み込みしてください。証明書をインストールするのが確実な解決方法です。',
      disconnectedNever:
        'キーボードとマウスを送る接続を開けませんでした。ページの他の部分はこの接続を使用しないため動作しています。お使いの環境とデバイスの間でこの接続がブロックされていないか確認してください。',
      disconnectedDropped:
        'キーボードとマウスを送る接続が切断され、復旧していません。再起動後は自動的に再接続します。この状態が続く場合は、ページを再読み込みしてください。',
      hidDisabled: 'このデバイスでは HID が無効です（/boot/disable_hid）。',
      keyFailed: 'キーを送信できませんでした。'
    },
    speaker: { title: 'スピーカー', unmute: 'ミュート解除', mute: 'ミュート' },
    menu: {
      collapse: 'メニューを折りたたむ',
      expand: 'メニューを展開する'
    },
    ion: {
      checking: 'ストリームを開始する前にビデオメモリを確認しています...',
      warn: 'ビデオメモリが不足しています。サーバーをあと 1 回再起動すると使い切ってしまいます。都合のよいときに再起動してください。',
      criticalTitle: 'ストリームを開始するためのビデオメモリが不足しています',
      criticalBody:
        'ビデオを開始すると予約メモリを使い切り、サーバーが停止します。電源制御や再起動を含む他のすべての機能は引き続き動作します。このメモリを回収できるのは NanoKVM の再起動だけです。',
      criticalContinue: 'それでもビデオを開始',
      criticalReboot: 'NanoKVM を再起動',
      criticalRebooting: '再起動中...'
    }
  }
};

export default ja;
