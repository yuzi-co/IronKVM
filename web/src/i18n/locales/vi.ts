const vi = {
  translation: {
    feedback: {
      enabled: 'Đã bật {{name}}',
      disabled: 'Đã tắt {{name}}',
      failed: 'Yêu cầu thất bại. Hãy thử lại.',
      network: 'Không kết nối được với thiết bị. Kiểm tra kết nối và thử lại.',
      saved: 'Đã lưu',
      timeout: 'Thiết bị phản hồi quá lâu. Hãy thử lại.'
    },
    common: {
      copy: 'Sao chép',
      copied: 'Đã sao chép',
      copyFailed: 'Không sao chép được. Hãy chọn văn bản và sao chép thủ công.',
      notUpdating: 'Không cập nhật: lần làm mới gần nhất thất bại.',
      off: 'Tắt',
      running: 'Đang chạy',
      save: 'Lưu',
      cancel: 'Hủy',
      delete: 'Xóa',
      remove: 'Gỡ bỏ'
    },
    head: {
      desktop: 'Màn hình từ xa',
      login: 'Đăng nhập',
      changePassword: 'Đổi mật khẩu',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Đã đổi mật khẩu. Hãy đăng nhập bằng mật khẩu mới.',
      cookieRejected:
        'Trình duyệt từ chối lưu phiên. Cookie còn sót lại từ phiên HTTPS trước không thể bị thay thế qua http thường. Hãy xóa cookie của địa chỉ này, hoặc mở cửa sổ ẩn danh, rồi đăng nhập lại.',
      login: 'Đăng nhập',
      placeholderUsername: 'Vui lòng nhập tên người dùng',
      placeholderPassword: 'vui lòng nhập mật khẩu',
      placeholderCurrentPassword: 'Mật khẩu hiện tại',
      placeholderPassword2: 'vui lòng nhập lại mật khẩu',
      noEmptyUsername: 'tên người dùng không được để trống',
      noEmptyPassword: 'mật khẩu không được để trống',
      passwordLength: 'Mật khẩu phải dài từ 8 đến 72 ký tự',
      noAccount:
        'Không thể lấy thông tin người dùng, vui lòng làm mới trang web hoặc đặt lại mật khẩu',
      invalidUser: 'tên người dùng hoặc mật khẩu không hợp lệ',
      locked: 'Đăng nhập quá nhiều, vui lòng thử lại sau',
      globalLocked: 'Hệ thống đang được bảo vệ, vui lòng thử lại sau',
      error: 'lỗi không mong đợi',
      invalidCurrentPassword: 'Mật khẩu hiện tại không đúng',
      changePassword: 'Đổi mật khẩu',
      changePasswordDesc: 'Để bảo mật thiết bị của bạn, vui lòng thay đổi mật khẩu đăng nhập web.',
      differentPassword: 'mật khẩu không khớp',
      illegalUsername: 'tên người dùng chứa ký tự không hợp lệ',
      illegalPassword: 'mật khẩu chứa ký tự không hợp lệ',
      forgetPassword: 'Quên mật khẩu',
      ok: 'OK',
      cancel: 'Hủy',
      loginButtonText: 'Đăng nhập',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the IronKVM for 10 seconds.',
        reset3: 'Tài khoản web mặc định:',
        reset4: 'Tài khoản SSH mặc định:',
        change1: 'Lưu ý rằng thao tác này sẽ thay đổi các mật khẩu sau:',
        change2: 'Mật khẩu đăng nhập web',
        change3: 'Mật khẩu root hệ thống (mật khẩu đăng nhập SSH)',
        change4: 'Để đặt lại mật khẩu, hãy nhấn và giữ nút BOOT trên IronKVM.',
        resetDocs: 'Xem các bước chi tiết trong tài liệu phần cứng:',
        hardwareDocs: 'Wiki Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Cấu hình Wi-Fi cho IronKVM',
      success: 'Please check the network status of IronKVM and visit the new IP address.',
      failed: 'Thao tác thất bại, vui lòng thử lại.',
      invalidMode:
        'Chế độ hiện tại không hỗ trợ thiết lập mạng. Vui lòng truy cập thiết bị của bạn và bật chế độ cấu hình Wi-Fi.',
      confirmBtn: 'Ok',
      finishBtn: 'Hoàn tất',
      ap: {
        authTitle: 'Yêu cầu xác thực',
        authDescription: 'Vui lòng nhập mật khẩu AP để tiếp tục',
        authFailed: 'Mật khẩu AP không hợp lệ',
        passPlaceholder: 'AP mật khẩu',
        verifyBtn: 'Xác minh'
      },
      ssidRequired: 'Nhập tên mạng, tối đa 32 ký tự',
      passwordLength: 'Mật khẩu dài 8 đến 63 ký tự. Để trống nếu là mạng mở.',
      passwordOptional: 'Mật khẩu (để trống nếu là mạng mở)',
      lost: 'Bo mạch đã ngừng phản hồi. Có thể nó đã vào mạng và tắt điểm phát cài đặt. Nếu điểm phát xuất hiện lại, việc kết nối đã thất bại: hãy kết nối lại và thử lại.',
      done: 'Đã cài đặt xong. Hãy kết nối lại thiết bị này với mạng thường dùng và mở bo mạch tại địa chỉ mới.'
    },
    screen: {
      codecNoWebrtcHevc: 'Trình duyệt này không nhận được H.265 qua WebRTC',
      codecNoHevc: 'Trình duyệt này không giải mã được H.265',
      codecNote:
        'Bo mạch chỉ có một bộ mã hóa nên thay đổi này áp dụng cho mọi người xem. Kết nối lại để áp dụng cho phiên WebRTC đang chạy.',
      codec: 'Codec',
      updateFailed: 'Chưa áp dụng được cài đặt',
      scale: 'Quy mô',
      title: 'Màn hình',
      video: 'Chế độ video',
      videoDirectTips: 'Bật HTTPS trong "Cài đặt > Thiết bị" để sử dụng chế độ này',
      resolution: 'Độ phân giải',
      ocr: {
        title: 'Đọc văn bản (OCR)',
        tips: 'Văn bản được nhận dạng trong trình duyệt này. Bạn có thể sửa trước khi sao chép.',
        hint: 'Kéo qua đoạn văn bản cần đọc. Nhấn Esc để hủy.',
        noPicture: 'Chờ video hiển thị, sau đó kéo qua đoạn văn bản cần đọc.',
        cancel: 'Hủy',
        language: 'Ngôn ngữ',
        languages: {
          eng: 'Tiếng Anh'
        },
        preview: 'Vùng đã chọn',
        capturing: 'Đang chụp màn hình...',
        loading: 'Đang tải tính năng nhận dạng văn bản...',
        recognizing: 'Đang đọc văn bản...',
        noText: 'Không tìm thấy văn bản trong vùng đã chọn.',
        copy: 'Sao chép',
        copied: 'Đã sao chép vào bộ nhớ tạm',
        copyFailed: 'Không thể sao chép vào bộ nhớ tạm',
        selectAgain: 'Chọn lại',
        unsupported:
          'Trình duyệt này không thể chạy nhận dạng văn bản. Tính năng này cần WebAssembly SIMD, có trên các trình duyệt hiện nay.',
        captureFailed: 'Không thể chụp màn hình.',
        outside: 'Vùng đã chọn nằm ngoài hình ảnh.',
        recognizeFailed: 'Nhận dạng văn bản thất bại.'
      },
      controlRegion: {
        title: 'Hiệu chỉnh chuột',
        description:
          'Sử dụng cài đặt này khi thiết bị được điều khiển dùng độ phân giải không phải tỷ lệ 16:9 và con trỏ bị lệch theo chiều ngang hoặc chiều dọc.',
        off: 'Tắt',
        auto: 'Tự động',
        autoWarning:
          'Hiệu chỉnh có thể thất bại khi ứng dụng của người dùng có nền hoàn toàn màu đen.',
        manual: 'Thủ công',
        selectedResolution: 'Độ phân giải vùng đã chọn',
        unused: 'Không sử dụng',
        originalResolution: 'Độ phân giải gốc',
        selectResolution: 'Chọn độ phân giải gốc',
        addResolution: 'Thêm độ phân giải tùy chỉnh',
        add: 'Thêm',
        duplicateResolution: 'Độ phân giải này đã tồn tại.',
        width: 'Chiều rộng',
        height: 'Chiều cao',
        apply: 'Tính toán và áp dụng',
        invalidResolution: 'Nhập độ phân giải gốc hợp lệ sau khi video sẵn sàng.',
        select: 'Chọn vùng',
        clear: 'Khôi phục tự động',
        saveFailed: 'Không thể lưu vùng đầu vào.',
        tooSmall: 'Vùng đã chọn quá nhỏ.',
        previewUnavailable: 'Không thể xem trước',
        clearConfirm: 'Khôi phục tính năng tự động phát hiện viền đen?',
        dragHint: 'Kéo để chọn vùng màn hình từ xa',
        finish: 'Xong',
        confirm: 'Xác nhận',
        cancel: 'Hủy'
      },
      auto: 'Tự động',
      autoTips:
        'Màn hình bị xé hoặc lệch chuột có thể xảy ra ở các độ phân giải nhất định. Hãy xem xét điều chỉnh độ phân giải của máy remote hoặc tắt chế độ tự động.',
      fps: 'FPS',
      customizeFps: 'Tùy chỉnh',
      quality: 'Chất lượng',
      qualityLossless: 'Tốt nhất',
      qualityHigh: 'Cao',
      qualityMedium: 'Trung bình',
      qualityLow: 'Thấp',
      frameDetect: 'Phát hiện khung hình',
      frameDetectTip:
        'Tính toán sự khác biệt giữa các khung hình. Dừng truyền video khi không có thay đổi trên màn hình máy chủ từ xa.',
      resetHdmi: 'Đặt lại HDMI',
      mixedH264: {
        title: 'Xung đột luồng H.264',
        description:
          'H.264 Direct và H.264 WebRTC đang được sử dụng đồng thời. Điều này có thể gây xé hình hoặc video bị hỏng. Vui lòng chỉ sử dụng một chế độ H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Kết nối WebRTC thất bại',
        description: 'Kiểm tra kết nối mạng hoặc chuyển đổi chế độ video.'
      },
      captureStatus: {
        hdmiError: 'Lỗi màn hình HDMI',
        unsupportedResolution: 'Độ phân giải hiện tại không được hỗ trợ',
        retrieving: 'Đang lấy màn hình...',
        changingResolution: 'Đang chuyển độ phân giải...',
        updateFailed: 'Hiện không thể cập nhật màn hình',
        videoError: 'Lỗi hiển thị video',
        noHdmi: 'Không phát hiện tín hiệu HDMI',
        unavailable: 'Hiện không thể hiển thị màn hình'
      },
      directConnectionFailed: 'Kết nối luồng video thất bại'
    },
    keyboard: {
      close: 'Đóng',
      title: 'Bàn phím',
      paste: 'Dán',
      tips: 'Gõ văn bản trên máy chủ dưới dạng các lần nhấn phím. Chọn bố cục bàn phím mà máy chủ đang dùng.',
      placeholder: 'Vui lòng nhập',
      submit: 'Gửi',
      virtual: 'Bàn phím',
      readClipboard: 'Đọc từ Clipboard',
      clipboardPermissionDenied:
        'Quyền bảng nhớ tạm bị từ chối. Vui lòng cho phép truy cập clipboard trong trình duyệt của bạn.',
      clipboardReadError: 'Không đọc được bảng nhớ tạm',
      mediaKeys: {
        title: 'Phím đa phương tiện',
        mute: 'Tắt tiếng',
        volumeDown: 'Giảm âm lượng',
        volumeUp: 'Tăng âm lượng',
        previous: 'Bài trước',
        playPause: 'Phát hoặc tạm dừng',
        next: 'Bài tiếp theo',
        stop: 'Dừng'
      },
      pasting: {
        layout: 'Bố cục bàn phím trên máy chủ',
        layouts: {
          us: 'Tiếng Anh (Mỹ)',
          uk: 'Tiếng Anh (Anh)',
          de: 'Tiếng Đức',
          fr: 'Tiếng Pháp',
          es: 'Tiếng Tây Ban Nha',
          it: 'Tiếng Ý',
          ptBr: 'Tiếng Bồ Đào Nha (Brazil)',
          se: 'Tiếng Thụy Điển / Phần Lan',
          ru: 'Tiếng Nga',
          ja: 'Tiếng Nhật',
          ko: 'Tiếng Hàn'
        },
        speed: 'Tốc độ gõ',
        speeds: {
          fast: 'Nhanh',
          normal: 'Bình thường',
          slow: 'Chậm'
        },
        estimate: 'Thời gian gõ: khoảng {{duration}}',
        untypeable: 'Ký tự mà bố cục này không gõ được: {{count}}',
        untypeableAt: 'dòng {{line}}, cột {{column}}',
        skipUntypeable: 'Gõ phần còn lại',
        shortcut: '{{shortcut}} gõ ngay nội dung bộ nhớ tạm lên máy chủ.',
        clipboardUnavailable:
          'Trình duyệt chỉ cho trang đọc bộ nhớ tạm qua HTTPS. Hãy dán văn bản vào ô bằng Ctrl+V.',
        clipboardEmpty: 'Bộ nhớ tạm không có văn bản.',
        tooLong: 'Văn bản quá dài. Giới hạn là {{max}} ký tự.',
        inProgress: 'Đang gõ một nội dung dán khác.',
        typing: 'Đang gõ trên máy chủ',
        done: 'Đã gõ xong',
        canceled: 'Đã hủy dán',
        failed: 'Dán thất bại',
        cancel: 'Hủy',
        controlBusy: 'Một bộ điều khiển khác đang dùng bàn phím.',
        hidError: 'Không gửi được các lần nhấn phím tới máy chủ.'
      },
      shortcut: {
        sendFailed: 'Chưa gửi: kết nối nhập liệu bị ngắt',
        title: 'Phím tắt',
        custom: 'Tùy chỉnh',
        capture: 'Bấm vào đây để chụp phím tắt',
        clear: 'Rõ ràng',
        save: 'Lưu',
        captureTips:
          'Việc ghi nhận các phím cấp hệ thống (như phím Windows) yêu cầu quyền toàn màn hình.',
        enterFullScreen: 'Chuyển sang chế độ toàn màn hình.'
      },
      leaderKey: {
        saveFailed: 'Không lưu được phím dẫn',
        title: 'Phím Leader',
        desc: 'Bỏ qua các hạn chế của trình duyệt và gửi các phím tắt hệ thống trực tiếp đến máy chủ từ xa.',
        howToUse: 'Cách sử dụng',
        simultaneous: {
          title: 'Chế độ đồng thời',
          desc1: 'Nhấn giữ phím Leader, rồi nhấn phím tắt.',
          desc2: 'Trực quan nhưng có thể xung đột với các phím tắt hệ thống.'
        },
        sequential: {
          title: 'Chế độ tuần tự',
          desc1: 'Nhấn phím Leader → nhấn phím tắt theo thứ tự → nhấn lại phím Leader.',
          desc2: 'Yêu cầu nhiều bước hơn nhưng hoàn toàn tránh được xung đột hệ thống.'
        },
        enable: 'Bật phím Leader',
        tip: 'Khi được gán làm phím Leader, phím này chỉ hoạt động như bộ kích hoạt phím tắt và mất hành vi mặc định.',
        placeholder: 'Vui lòng nhấn phím Leader',
        shiftRight: 'Shift phải',
        ctrlRight: 'Ctrl phải',
        metaRight: 'Win phải',
        submit: 'Gửi',
        recorder: {
          rec: 'REC',
          activate: 'Kích hoạt phím',
          input: 'Hãy nhấn phím tắt...'
        }
      }
    },
    mouse: {
      title: 'Chuột',
      cursor: 'Kiểu con trỏ',
      default: 'Con trỏ mặc định',
      pointer: 'Con trỏ trỏ',
      cell: 'Con trỏ ô',
      text: 'Con trỏ văn bản',
      grab: 'Con trỏ nắm',
      hide: 'Ẩn con trỏ',
      mode: 'Chế độ chuột',
      absolute: 'Chế độ tuyệt đối',
      relative: 'Chế độ tương đối',
      absoluteShort: 'Tuyệt đối',
      relativeShort: 'Tương đối',
      touch: 'Chế độ cảm ứng',
      touchShort: 'Cảm ứng',
      absoluteStalled: 'Máy đích đang bỏ qua chuột tuyệt đối',
      absoluteStalledDesc:
        'Máy đích đã ngừng nhận báo cáo chuột tuyệt đối, nên các chuyển động con trỏ bị mất. Bàn phím không bị ảnh hưởng. Khôi phục USB thường khắc phục được; chế độ tương đối dùng một endpoint khác.',
      useRelative: 'Chuyển sang chế độ tương đối',
      direction: 'Hướng bánh xe cuộn',
      scrollUp: 'Giống máy tính này',
      scrollDown: 'Đảo ngược (cuộn tự nhiên)',
      speed: 'Tốc độ bánh xe cuộn',
      fast: 'Nhanh lên',
      slow: 'Chậm',
      requestPointer:
        'Đang sử dụng chế độ tương đối. Vui lòng nhấp vào màn hình để lấy con trỏ chuột.',
      resetHid: 'Đặt lại HID',
      hidOnly: {
        switchFailed: 'Không chuyển được chế độ. Kiểm tra kết nối rồi thử lại.',
        title: 'Chế độ chỉ HID',
        desc: 'Nếu chuột và bàn phím của bạn ngừng phản hồi và việc đặt lại HID không có tác dụng thì đó có thể là sự cố tương thích giữa IronKVM và thiết bị. Hãy thử bật chế độ HID-Only để tương thích tốt hơn.',
        tip1: 'Kích hoạt HID-Chế độ chỉ sẽ ngắt kết nối đĩa U ảo và mạng ảo',
        tip2: 'Ở chế độ HID-Chỉ, tính năng gắn hình ảnh bị tắt',
        rebuild: 'Chuyển chế độ sẽ dựng lại kết nối USB. IronKVM không khởi động lại',
        enable: 'Bật chế độ HID-Chỉ',
        disable: 'Tắt chế độ HID-Chỉ'
      },
      resetHidDone: 'Đã đặt lại USB HID',
      resetHidFailed: 'Đặt lại USB HID thất bại'
    },
    image: {
      delete: 'Xóa',
      inUse: 'Đang dùng. Hãy đẩy ra trước khi xóa.',
      retry: 'Thử lại',
      loadFailed: 'Không tải được danh sách ảnh đĩa',
      readOnlyLocked: 'Hãy đẩy đĩa ra để thay đổi. Cài đặt áp dụng khi chèn ảnh đĩa.',
      title: 'Hình ảnh',
      loading: 'Đang tải...',
      empty: 'Không tìm thấy',
      mountMode: 'Chế độ gắn kết',
      mountFailed: 'Mount thất bại',
      mountDesc: 'Trong một số hệ thống, cần phải eject đĩa ảo trên máy remote trước khi mount.',
      unmountFailed: 'Tháo lắp không thành công',
      unmountDesc:
        'Trên một số hệ thống, bạn cần đẩy hình ảnh ra khỏi máy chủ từ xa theo cách thủ công trước khi ngắt kết nối hình ảnh.',
      refresh: 'Làm mới danh sách hình ảnh',
      disk: 'Đĩa',
      cdrom: 'CD',
      driveEmpty: 'Trống',
      eject: 'Đẩy ra',
      readOnly: 'Chỉ đọc',
      readOnlyTip: 'Áp dụng cho hình ảnh tiếp theo được đưa vào đĩa.',
      noDrives: 'Không có ổ đĩa ảo. Hãy bật đĩa ảo trong Cài đặt.',
      insertFailed: 'Đưa vào thất bại',
      ejectFailed: 'Đẩy ra thất bại',
      insertInto: 'Đưa vào {{drive}}. Nhấp để thay đổi.',
      loadedIn: 'Trong ổ {{drive}}',
      attention: 'Chú ý',
      deleteConfirm: 'Bạn có chắc chắn muốn xóa hình ảnh này không?',
      okBtn: 'Có',
      cancelBtn: 'Không',
      deleteFailed: 'Xóa thất bại',
      ventoy: {
        statusNoKernel: 'Firmware này không hỗ trợ',
        statusNotInstalled: 'Chưa cài đặt',
        statusReady: 'Sẵn sàng',
        statusSelected: 'Ảnh đĩa đã chọn: {{count}}',
        statusInDrive: 'Trong ổ đĩa, {{size}}',
        noKernel:
          'Nhân của firmware này không hỗ trợ device-mapper, nên không thể dùng Ventoy cho đến khi cài một bản có hỗ trợ.',
        installDesc: 'Khởi động máy chủ từ nhiều ảnh đĩa trên một đĩa, không cần sao chép.',
        install: 'Cài đặt',
        installing: 'Đang tải Ventoy, khoảng 20 MB. Việc này có thể mất vài phút.',
        needsData: 'Ventoy cần bản IronKVM có phân vùng /data đã được gắn.',
        uninstall: 'Gỡ cài đặt',
        uninstallConfirm: 'Xóa các tệp của Ventoy?',
        noImages: 'Không có ảnh đĩa nào để đưa vào đĩa Ventoy.',
        onDisk: 'Trên đĩa Ventoy',
        missing: 'Thiếu: {{file}}',
        remove: 'Bỏ khỏi đĩa Ventoy',
        setHint: 'Chỉ có thể thay đổi bộ ảnh đĩa khi đĩa Ventoy không nằm trong ổ nào.',
        useAsDisk: 'Dùng làm đĩa ảo',
        failed: 'Yêu cầu Ventoy thất bại',
        secureBoot:
          'Khi bật Secure Boot, máy chủ phải đăng ký khóa của Ventoy trong MokManager một lần. Tệp khóa ENROLL_THIS_KEY_IN_MOKMANAGER.cer nằm trên phân vùng VTOYEFI.',
        readOnly:
          'Máy chủ thấy đĩa ở chế độ chỉ đọc, nên tính năng lưu trữ bền vững của Ventoy và ventoy.json trên ổ không hoạt động.'
      },
      tips: {
        title: 'Cách tải lên',
        usb1: 'Kết nối IronKVM với máy tính của bạn qua USB.',
        usb2: 'Đảm bảo rằng đĩa ảo đã được gắn kết (Cài đặt - Đĩa ảo).',
        usb3: 'Mở đĩa ảo trên máy tính của bạn và sao chép vào thư mục gốc của đĩa ảo.',
        scp1: 'Đảm bảo IronKVM và máy tính của bạn đang trên cùng một mạng nội bộ.',
        scp2: 'Mở terminal trên máy tính và sử dụng lệnh SCP để tải đĩa ảo lên thư mục /data trên IronKVM.',
        scp3: 'Ví dụ: scp đường-dẫn-image root@ip-của-nanokvm:/data',
        tfCard: 'Thẻ TF',
        tf1: 'Phương pháp này được hỗ trợ trên hệ thống Linux',
        tf2: 'Lấy thẻ TF từ IronKVM (với phiên bản FULL, hãy tháo vỏ trước).',
        tf3: 'Chèn thẻ TF vào đầu đọc thẻ và kết nối với máy tính của bạn.',
        tf4: 'Sao chép tệp hình ảnh vào thư mục /data trên thẻ TF.',
        tf5: 'Chèn thẻ TF vào IronKVM.'
      }
    },
    script: {
      title: 'Script',
      upload: 'Tải lên',
      run: 'Chạy',
      runBackground: 'Chạy nền',
      runFailed: 'Chạy thất bại',
      attention: 'Chú ý',
      delDesc: 'Bạn có chắc chắn muốn xóa tệp này không?',
      confirm: 'Có',
      cancel: 'Không',
      delete: 'Xóa',
      close: 'Đóng',
      empty: 'Chưa có script nào. Tải lên tệp .sh hoặc .py để chạy trên bo mạch.',
      loadFailed: 'Không thể tải danh sách script',
      uploaded: 'Đã tải lên script',
      uploadFailed: 'Không thể tải lên script',
      started: 'Đã chạy script trong nền',
      deleteFailed: 'Không thể xóa script',
      waitLimit: 'Đang chờ script chạy xong, tối đa {{minutes}} phút.',
      timedOut:
        'Script chạy lâu hơn {{minutes}} phút nên trang này đã ngừng chờ. Có thể script vẫn đang chạy trên bo mạch.'
    },
    terminal: {
      invalidBaud: 'Tốc độ baud này không được hỗ trợ.',
      invalidPort: 'Nhập đường dẫn thiết bị trong /dev, ví dụ /dev/ttyS1.',
      invalidSettings: 'Cài đặt cổng nối tiếp không hợp lệ. Đây là shell của chính bo mạch.',
      disconnected: 'Đã ngắt kết nối. Nhấn Enter để kết nối lại.',
      title: 'Terminal',
      nanokvm: 'Terminal IronKVM',
      serial: 'Terminal Cổng Nối Tiếp',
      serialPort: 'Cổng Nối Tiếp',
      serialPortPlaceholder: 'Vui lòng nhập cổng nối tiếp',
      baudrate: 'Tốc độ Baud',
      parity: 'Tính chẵn lẻ',
      parityNone: 'Không có',
      parityEven: 'Chẵn',
      parityOdd: 'Lẻ',
      flowControl: 'Kiểm soát luồng',
      flowControlNone: 'Không có',
      flowControlSoft: 'Phần mềm',
      flowControlHard: 'Phần cứng',
      dataBits: 'Bit dữ liệu',
      stopBits: 'Bit dừng',
      confirm: 'OK'
    },
    wol: {
      no: 'Không',
      yes: 'Có',
      deleteConfirm: 'Xóa địa chỉ đã lưu này?',
      delete: 'Xóa',
      wake: 'Đánh thức',
      rename: 'Đổi tên',
      showMac: 'Hiện địa chỉ MAC',
      showName: 'Hiện tên',
      requestFailed: 'Không kết nối được tới thiết bị để gửi lệnh',
      deleteFailed: 'Không xóa được',
      renameFailed: 'Không đổi tên được',
      title: 'Wake-on-LAN',
      sending: 'Đang gửi lệnh...',
      sent: 'Đã gửi lệnh',
      input: 'Vui lòng nhập địa chỉ MAC',
      ok: 'OK'
    },
    download: {
      uploadFailed: 'Tải lên thất bại',
      uploadSuccess: 'Đã tải lên xong',
      uploading: 'Đang tải lên: {{file}}',
      downloadingPercent: 'Đang tải xuống ({{percent}}): {{file}}',
      downloading: 'Đang tải xuống: {{file}}',
      title: 'Trình tải xuống hình ảnh',
      input: 'Vui lòng nhập hình ảnh từ xa URL',
      ok: 'OK',
      disabled: '/data phân vùng là RO nên không tải được image',
      uploadbox: 'Thả file vào đây hoặc bấm vào để chọn',
      inputfile: 'Vui lòng nhập File hình ảnh',
      NoISO: 'Không có ISO',
      sha256: 'SHA-256 (tùy chọn)',
      sha256Placeholder: 'Nhập mã kiểm tra SHA-256 gồm 64 ký tự',
      invalidSHA256: 'SHA-256 phải là chuỗi thập lục phân gồm 64 ký tự',
      failed: 'Tải xuống thất bại',
      success: 'Tải xuống thành công',
      checksumFailed: 'Tải xuống thất bại: xác minh SHA-256 không thành công',
      cancel: 'Hủy',
      cancelFailed: 'Không thể hủy tải xuống',
      bootMenu: 'Menu khởi động (netboot.xyz)',
      bootMenuPresent: '{{file}} đã có trên thiết bị, checksum đúng',
      bootMenuDesc: 'Tải ISO netboot.xyz, đã kiểm tra checksum, cho CD ảo'
    },
    power: {
      title: 'Nguồn',
      showConfirm: 'Xác nhận',
      showConfirmTip: 'Hỏi trước khi nhấn nguồn ngắn. Reset và nhấn giữ luôn hỏi.',
      reset: 'Đặt lại',
      power: 'Nguồn',
      powerShort: 'Nguồn (nhấp ngắn)',
      powerLong: 'Nguồn (nhấp dài)',
      resetConfirm: 'Tiến hành thao tác đặt lại?',
      powerConfirm: 'Tiếp tục vận hành nguồn điện?',
      okBtn: 'Có',
      cancelBtn: 'Không',
      hostOs: 'HĐH của máy chủ',
      hostOsTip: 'Được gửi dưới dạng phím USB. Máy chủ quyết định chúng làm gì.',
      sleep: 'Ngủ',
      wake: 'Đánh thức',
      wakeKey: 'Đánh thức bằng Shift',
      powerDown: 'Tắt máy',
      sleepConfirm: 'Cho máy chủ vào chế độ ngủ?',
      powerDownConfirm: 'Gửi phím tắt nguồn đến máy chủ?',
      wakeTip:
        'Máy chủ đang ngủ thường bỏ qua lệnh Đánh thức từ thiết bị đã cho nó ngủ. Đánh thức bằng Shift nhấn một phím trên bàn phím, điều mà nhiều máy chủ chấp nhận hơn.',
      led: 'Đèn LED nguồn',
      ledOn: 'Sáng',
      ledOff: 'Tắt',
      ledUnknown: 'Không xác định',
      ledConnected: 'Đã nối đèn LED nguồn',
      ledConnectedTip:
        'Chỉ bật khi chân cắm đèn LED nguồn của máy chủ được nối dây vào bo mạch. Nếu không, trạng thái nguồn sẽ không xác định.',
      ledConnectedFailed: 'Không lưu được cài đặt đèn LED nguồn',
      powerLongConfirm: 'Giữ nút nguồn {{seconds}} giây? Thao tác này ngắt điện mà không tắt máy.',
      done: 'Đã nhấn nút',
      failed: 'Nhấn nút thất bại'
    },
    settings: {
      title: 'Cài đặt',
      nav: {
        general: 'Chung',
        device: 'Thiết bị',
        network: 'Mạng',
        remote: 'Truy cập từ xa',
        boot: 'Khởi động',
        locked: 'Đang có thao tác chạy. Không thể chuyển trang hoặc đóng cho đến khi hoàn tất.',
        vpnProvider: 'Nhà cung cấp VPN'
      },
      mcp: {
        keyNote:
          'MCP dùng khóa API riêng, hiển thị bên dưới. Khóa từ trang Khóa API không dùng được ở đây.',
        title: 'Dịch vụ MCP',
        service: 'Điều khiển từ xa MCP',
        serviceDesc:
          'Cho phép các máy khách MCP đáng tin cậy điều khiển bàn phím và chuột và chụp ảnh màn hình',
        securityWarning:
          'Bất kỳ ai có khóa API này đều có thể điều khiển máy chủ từ xa và xem màn hình. Hãy sử dụng HTTPS và chỉ bật dịch vụ trên các mạng đáng tin cậy.',
        endpoint: 'Điểm cuối',
        apiKey: 'Khóa API',
        regenerateConfirmTitle: 'Tạo lại khóa API MCP?',
        regenerateConfirmDesc: 'Khóa hiện tại sẽ ngừng hoạt động ngay lập tức.',
        enableConfirmTitle: 'Bật điều khiển MCP bên ngoài?',
        enableConfirmDesc: 'Bật MCP sẽ dừng PicoClaw và đóng tất cả phiên PicoClaw đang hoạt động.',
        failed: 'Thao tác MCP không thành công',
        copyFailed: 'Sao chép thất bại. Vui lòng sao chép thủ công.',
        okBtn: 'Xác nhận',
        cancelBtn: 'Hủy',
        showKey: 'Hiện khóa',
        hideKey: 'Ẩn khóa',
        regenerateKey: 'Tạo lại khóa'
      },
      redfish: {
        example: 'Ví dụ',
        title: 'Redfish',
        service: 'Dịch vụ Redfish',
        serviceDesc:
          'API Redfish của DMTF, dùng để điều khiển nguồn, phương tiện ảo và xem trạng thái từ các công cụ như redfishtool và Ansible. Tắt dịch vụ sẽ kết thúc mọi phiên Redfish.',
        endpoint: 'Gốc dịch vụ',
        httpsOn: 'Bo mạch đang phục vụ HTTPS, điều mà hầu hết công cụ Redfish cần.',
        httpsOff:
          'Bo mạch đang phục vụ HTTP thường. Hầu hết công cụ Redfish cần HTTPS: hãy bật nó trong "Cài đặt > Mạng".',
        credentials:
          'Redfish dùng các tài khoản KVM, với xác thực Basic hoặc phiên Redfish, và khóa API gửi dưới dạng X-Auth-Token. Khóa API được quản lý trên trang Khóa API.',
        powerActions: 'Thao tác nguồn',
        powerActionsDesc:
          'Các kiểu reset hiện có. On, ForceOff và GracefulShutdown cần trạng thái nguồn, nên chỉ có khi "Đã nối đèn LED nguồn" được bật trong menu nguồn.',
        sessions: 'Phiên',
        noSessions: 'Không có phiên Redfish nào đang mở',
        created: 'Tạo lúc',
        lastUsed: 'Dùng lần cuối',
        refresh: 'Làm mới',
        end: 'Kết thúc',
        endConfirmTitle: 'Kết thúc phiên Redfish này?',
        endConfirmDesc:
          'Token của phiên sẽ ngừng hoạt động ngay lập tức. Máy khách phải đăng nhập lại.',
        failed: 'Thao tác Redfish thất bại',
        copyFailed: 'Sao chép thất bại. Vui lòng sao chép thủ công.',
        okBtn: 'Xác nhận',
        cancelBtn: 'Hủy'
      },
      ipmi: {
        copyBeforeSave: 'Hãy sao chép mật khẩu ngay. Sau khi lưu sẽ không thể hiển thị lại.',
        noLogin:
          'IPMI đang bật, nhưng không tài khoản đang hoạt động nào có mật khẩu IPMI, nên không ai đăng nhập được. Hãy đặt một mật khẩu bên dưới.',
        title: 'IPMI',
        warning:
          'Xác thực IPMI vốn yếu do thiết kế. Bất kỳ ai truy cập được bo mạch và biết một tên người dùng đều có thể lấy hash mật khẩu IPMI của người dùng đó và thử bẻ khóa ngoại tuyến. Hãy dùng mật khẩu được tạo tự động, chỉ bật IPMI trong mạng tin cậy và ưu tiên Redfish qua HTTPS khi công cụ hỗ trợ.',
        service: 'IPMI qua LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) trên cổng UDP 623, cho nguồn và trạng thái của máy chủ. IPMI 1.5 và cipher suite 0 bị từ chối. Tắt đi sẽ kết thúc mọi phiên IPMI.',
        example: 'Ví dụ',
        copyFailed: 'Sao chép thất bại. Hãy sao chép thủ công.',
        ledOn: 'Có sẵn trạng thái nguồn, on, off, soft, cycle và reset.',
        ledOff:
          '"Đã nối đèn LED nguồn" đang tắt trong menu nguồn, nên không biết trạng thái nguồn. Chỉ "power reset" hoạt động: status, on, off, soft và cycle bị từ chối.',
        accounts: 'Tài khoản',
        accountsDesc:
          'IPMI đăng nhập bằng các tài khoản KVM, mỗi tài khoản có mật khẩu IPMI riêng, tách biệt với mật khẩu web. Quản trị viên nhận ADMINISTRATOR. Người dùng nhận USER: họ có thể đọc trạng thái nguồn với "-L USER" nhưng không thể thay đổi.',
        passwordSet: 'Đã đặt mật khẩu IPMI',
        passwordNotSet: 'Chưa có mật khẩu IPMI: không thể đăng nhập qua IPMI',
        nameTooLong: 'Tên dài hơn 16 ký tự, điều mà IPMI không cho phép',
        accountDisabled: 'Tài khoản đã bị vô hiệu hóa',
        setPassword: 'Đặt mật khẩu',
        changePassword: 'Đổi mật khẩu',
        remove: 'Xóa',
        removeConfirmTitle: 'Xóa mật khẩu IPMI của {{user}}?',
        removeConfirmDesc:
          'Tài khoản sẽ không thể đăng nhập qua IPMI nữa, và các phiên IPMI của nó sẽ kết thúc.',
        passwordTitle: 'Mật khẩu IPMI cho {{user}}',
        passwordDesc:
          'Từ 12 đến 20 ký tự ASCII in được, khác với mật khẩu web. IPMI yêu cầu bo mạch lưu mật khẩu ở dạng có thể đọc lại, vì vậy hãy dùng mật khẩu không dùng ở nơi nào khác. Hãy sao chép trước khi lưu: mật khẩu sẽ không hiển thị lại.',
        passwordPlaceholder: 'Mật khẩu IPMI',
        generate: 'Tạo',
        copy: 'Sao chép',
        save: 'Lưu',
        passwordLength: 'Dùng từ 12 đến 20 ký tự.',
        passwordChars: 'Chỉ dùng ký tự ASCII in được.',
        saved: 'Đã lưu mật khẩu IPMI',
        failed: 'Thao tác IPMI thất bại',
        okBtn: 'Xác nhận',
        cancelBtn: 'Hủy'
      },
      ssh: {
        service: 'Máy chủ SSH',
        serviceDesc: 'Chạy sshd ngay bây giờ và mỗi lần khởi động',
        failed: 'Không tải được cài đặt SSH',
        rootDefault: 'root vẫn dùng mật khẩu mặc định của nhà sản xuất',
        rootEmpty: 'root chưa có mật khẩu',
        rootWarning:
          'Bất kỳ ai truy cập được console hoặc SSH đều có thể đăng nhập bằng root. Đặt mật khẩu tại {{account}} > {{password}}: với chủ thiết bị, thao tác này cũng đặt mật khẩu root.',
        connection: 'Kết nối',
        command: 'Đăng nhập bằng root',
        port: 'Cổng',
        viaVpn: 'Qua {{name}}',
        notRunning: 'sshd không chạy. Bật máy chủ SSH để kết nối.',
        hostKeys: 'Dấu vân tay khóa máy chủ',
        hostKeysDesc: 'Đối chiếu với thông tin ssh hiển thị ở lần kết nối đầu tiên.',
        noHostKeys: 'Chưa có khóa máy chủ. sshd tạo chúng khi khởi động lần đầu.',
        keys: 'Khóa được ủy quyền',
        keysDesc:
          'Các khóa công khai có thể đăng nhập bằng root. Chúng được lưu trên phân vùng dữ liệu nên vẫn còn sau khi cập nhật.',
        noKeys: 'Chưa có khóa được ủy quyền.',
        noComment: 'không có ghi chú',
        addPlaceholder: 'Dán một khóa công khai, ví dụ nội dung của ~/.ssh/id_ed25519.pub',
        add: 'Thêm khóa',
        added: 'Đã thêm khóa',
        removed: 'Đã xóa khóa',
        deleteConfirm: 'Xóa khóa này?',
        deleteConfirmDesc: 'Khóa này sẽ không thể đăng nhập nữa. Các phiên đang mở vẫn giữ nguyên.',
        invalidKey: 'Đây không phải khóa công khai. Hãy dán một dòng duy nhất từ tệp .pub.',
        keyOptions: 'Không chấp nhận khóa có tùy chọn như command= hoặc from= ở đây.',
        duplicateKey: 'Khóa này đã được ủy quyền.',
        lastKey: 'Không thể xóa khóa cuối cùng khi chế độ chỉ dùng khóa đang bật.',
        keysOnly: 'Chỉ dùng khóa',
        keysOnlyDesc:
          'Tắt đăng nhập bằng mật khẩu và keyboard-interactive. Các phiên đang mở vẫn giữ nguyên.',
        keysOnlyNeedsKey:
          'Hãy thêm một khóa được ủy quyền trước, nếu không sẽ không ai đăng nhập được.',
        keysOnlyOn: 'Đã tắt đăng nhập bằng mật khẩu',
        keysOnlyOff: 'Đã bật đăng nhập bằng mật khẩu',
        notHonoured:
          'sshd trong image này không đọc cài đặt này, nên đăng nhập bằng mật khẩu vẫn bật.',
        reloadFailed:
          'Đã lưu, nhưng không tải lại được sshd. Cài đặt sẽ áp dụng khi sshd khởi động lần tới.',
        notApplied: 'sshd vẫn chấp nhận mật khẩu. Tắt rồi bật lại máy chủ SSH để áp dụng cài đặt.'
      },
      vnc: {
        address: 'Địa chỉ',
        certHint:
          'VeNCrypt X509Plain dùng chứng chỉ tự ký của thiết bị, nên máy khách sẽ cảnh báo ở lần kết nối đầu. Hãy chấp nhận, hoặc lưu chứng chỉ từ địa chỉ HTTPS của trang này và đưa cho TigerVNC bằng -X509CA=<tệp>.',
        example: 'Ví dụ',
        title: 'VNC',
        service: 'Máy chủ VNC',
        serviceDesc:
          'Cho phép một trình khách VNC, như TigerVNC hoặc Remmina, xem và điều khiển máy chủ đích. Trình khách phải hỗ trợ mã hóa Tight. Mỗi lần một phiên.',
        credentials:
          'Đăng nhập bằng tài khoản KVM. Kết nối được mã hóa bằng chứng chỉ TLS của bo mạch (VeNCrypt X509Plain).',
        port: 'Cổng',
        portDesc: 'Cổng TCP mà máy chủ lắng nghe.',
        maxFps: 'Giới hạn tốc độ khung hình',
        maxFpsDesc: 'Số khung hình mỗi giây tối đa gửi cho một trình khách.',
        vncAuth: 'Xác thực VNC đơn giản',
        vncAuthDesc:
          'Dành cho trình khách không có VeNCrypt. Cách này kiểm tra một mật khẩu VNC riêng thay vì tài khoản.',
        vncAuthWarning:
          'Xác thực VNC đơn giản không mã hóa kết nối. Bất kỳ ai trên đường mạng đều có thể thấy màn hình và các phím đã gõ. Chỉ dùng trên mạng tin cậy.',
        password: 'Mật khẩu VNC',
        passwordSet: 'Đã đặt mật khẩu. Nhập mật khẩu mới để thay đổi.',
        passwordInvalid: 'Mật khẩu VNC phải có từ 6 đến 8 ký tự.',
        save: 'Lưu',
        saved: 'Đã lưu cài đặt',
        state: 'Trạng thái',
        listening: 'Đang lắng nghe trên cổng {{port}}',
        notListening: 'Không lắng nghe',
        noSession: 'Không có phiên nào đang mở',
        client: 'Trình khách',
        user: 'Người dùng',
        method: 'Xác thực',
        methodVencrypt: 'Tài khoản qua TLS',
        methodVnc: 'Mật khẩu VNC',
        since: 'Đã kết nối từ',
        resolution: 'Độ phân giải',
        framesSent: 'Khung hình đã gửi',
        lastError: 'Phiên gần nhất đã kết thúc: {{error}}',
        refresh: 'Làm mới',
        disconnect: 'Ngắt kết nối',
        disconnectConfirmTitle: 'Kết thúc phiên VNC?',
        disconnectConfirmDesc:
          'Trình khách bị ngắt ngay, và mọi phím và nút đang giữ đều được nhả ra.',
        failed: 'Thao tác VNC thất bại',
        okBtn: 'Xác nhận',
        cancelBtn: 'Hủy'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Watchdog máy chủ',
        serviceDesc:
          'Nếu máy chủ lẽ ra đang chạy mà hình ảnh không thay đổi, hoặc không có tín hiệu HDMI, trong suốt thời gian chờ, bo mạch sẽ nhấn reset hoặc tắt rồi bật lại máy chủ.',
        stillWarning:
          'Máy chủ có màn hình chuyển sang chế độ ngủ, hoặc có hình ảnh đứng yên khi đang làm việc, trông giống như bị treo. Hãy tắt chế độ ngủ màn hình trên máy chủ, hoặc đặt một địa chỉ ping.',
        ledHint:
          '"Đã nối đèn LED nguồn" đang tắt trong menu nguồn. Watchdog không biết khi nào máy chủ tắt, nên coi máy chủ luôn bật.',
        timeout: 'Thời gian chờ',
        timeoutDesc:
          'Khoảng thời gian máy chủ có thể không có dấu hiệu hoạt động trước khi watchdog can thiệp.',
        action: 'Hành động',
        actionDesc: 'Tắt rồi bật giữ nút nguồn trong 5 giây, sau đó nhấn lại.',
        actionReset: 'Đặt lại',
        actionPower: 'Tắt rồi bật',
        cooldown: 'Thời gian nghỉ',
        cooldownDesc: 'Khoảng thời gian ngắn nhất giữa hai hành động.',
        maxPerHour: 'Hành động mỗi giờ',
        maxPerHourDesc: 'Số hành động tối đa trong một giờ.',
        pingHost: 'Địa chỉ ping',
        pingHostDesc:
          'Địa chỉ IP của máy chủ. Một phản hồi được tính là dấu hiệu hoạt động. Để trống nếu không ping.',
        pingHostInvalid: 'Nhập địa chỉ IPv4 hoặc IPv6.',
        minutes: 'phút',
        save: 'Lưu',
        saved: 'Đã lưu',
        state: 'Bộ phát hiện',
        status: {
          off: 'Tắt',
          watching: 'Đang theo dõi',
          hostOff: 'Máy chủ đã tắt',
          captureOff: 'Đã tắt thu HDMI',
          cooldown: 'Đang nghỉ',
          capped: 'Đã đạt giới hạn mỗi giờ',
          acting: 'Đang can thiệp'
        },
        signal: 'Tín hiệu HDMI',
        yes: 'Có',
        no: 'Không',
        led: 'Đèn LED nguồn',
        on: 'Sáng',
        off: 'Tắt',
        ledNotConnected: 'Chưa kết nối',
        ping: 'Ping',
        pingNotSet: 'Chưa đặt',
        pingReply: 'Có phản hồi',
        pingNoReply: 'Không phản hồi',
        lastChange: 'Lần thay đổi hình ảnh gần nhất',
        never: 'Chưa bao giờ',
        actsIn: 'Can thiệp sau',
        actionsLastHour: 'Hành động trong giờ qua',
        duration: '{{minutes}} phút {{seconds}} giây',
        log: 'Nhật ký',
        noLog: 'Watchdog chưa can thiệp lần nào.',
        refresh: 'Làm mới',
        reasonFrozen: 'Hình ảnh không thay đổi',
        reasonNoSignal: 'Không có tín hiệu HDMI',
        stuckFor: 'không có dấu hiệu hoạt động trong {{duration}}',
        pressFailed: 'Nhấn nút thất bại: {{error}}',
        noScreenshot: 'Không có ảnh chụp màn hình',
        failed: 'Thao tác watchdog thất bại',
        powerNeedsLed: 'Chu kỳ nguồn cần bật "Đã nối đèn LED nguồn" trong menu nguồn.',
        noLedConfirmTitle: 'Bật watchdog khi không có đèn LED nguồn?',
        noLedConfirmDesc:
          'Bo mạch không thấy khi nào máy chủ tắt, nên coi máy chủ luôn bật. Nếu bạn tắt máy chủ, watchdog sẽ nhấn reset khi hết thời gian chờ. Hãy nối đèn LED nguồn để tránh điều này.',
        noLedConfirmOk: 'Bật',
        cancel: 'Hủy'
      },
      netboot: {
        title: 'Khởi động qua mạng',
        description:
          'Khởi động máy chủ từ mạng: iPXE và menu các ảnh đĩa trên KVM qua liên kết mạng USB, hoặc netboot.xyz qua proxy DHCP trên LAN.',
        addon: 'dnsmasq và tệp khởi động',
        addonDesc:
          'Cài vào /data: dnsmasq từ Alpine, iPXE và netboot.xyz từ bản phát hành của chúng, mỗi tệp được kiểm tra theo checksum.',
        install: 'Cài đặt',
        installing: 'Đang cài đặt. Việc này có thể mất vài phút.',
        uninstall: 'Gỡ cài đặt',
        uninstallConfirm: 'Tắt khởi động qua mạng và gỡ dnsmasq cùng các tệp khởi động?',
        needsData: 'Khởi động qua mạng cần ảnh IronKVM có phân vùng /data được gắn.',
        usb: 'Trên liên kết mạng USB',
        usbDesc:
          'Khi liên kết mạng USB bật, dnsmasq phục vụ nó thay cho udhcpd. Máy chủ nhận địa chỉ duy nhất của nó mà không có bộ định tuyến và không có máy chủ DNS, iPXE cho kiến trúc của nó và menu các ảnh ISO trên KVM.',
        linkOff: 'Liên kết mạng USB đang tắt. Hãy bật nó trong Thiết bị, Mạng USB.',
        menuUrl: 'Menu',
        leases: 'Hợp đồng thuê của máy chủ',
        noLeases: 'Chưa có',
        netbootxyzNote:
          'netboot.xyz trong menu tải từ internet, nơi liên kết USB không tới được. Máy chủ cần internet trên một cổng mạng khác.',
        lan: 'Proxy DHCP trên LAN',
        lanDesc:
          'Trả lời các máy khách PXE trên LAN bằng netboot.xyz, sau đó netboot.xyz tải menu từ internet. Nó không bao giờ cấp địa chỉ và không phục vụ các ảnh đĩa trên KVM.',
        lanWarning:
          'Mọi máy khách PXE trên LAN này đều được mời dùng netboot.xyz, không chỉ máy chủ. Chỉ bật tính năng này trên mạng bạn kiểm soát.',
        lanConfirm: 'Bật proxy DHCP trên LAN?',
        lanInterface: 'LAN',
        running: 'Đang chạy',
        stopped: 'Không chạy',
        images: 'Ảnh đĩa trong menu',
        noImages: 'Không có ảnh ISO trong thư mục ảnh.',
        boots: 'Lần khởi động gần đây',
        noBoots: 'Máy chủ chưa tải gì.',
        log: 'Nhật ký dnsmasq',
        refresh: 'Làm mới',
        okBtn: 'Xác nhận',
        cancelBtn: 'Hủy',
        failed: 'Thao tác khởi động qua mạng thất bại'
      },
      about: {
        title: 'Giới thiệu về IronKVM',
        information: 'Thông tin',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Phiên bản Ứng dụng',
        applicationTip: 'Phiên bản ứng dụng web IronKVM',
        image: 'Phiên bản Hình ảnh',
        imageTip: 'Ảnh thẻ IronKVM và ảnh hệ thống NanoKVM mà nó dựa trên',
        kernel: 'Phiên bản Kernel',
        kernelTip: 'Bản phát hành của kernel Linux đang chạy',
        deviceKey: 'Khóa Thiết bị',
        videoMemory: 'Bộ nhớ video',
        videoMemoryTip:
          'Bộ nhớ dành riêng cho việc thu video. Nó không được chia sẻ với phần còn lại của hệ thống.',
        videoMemoryGenerations_other: '{{count}} phiên IronKVM trước đó đang giữ bộ nhớ video',
        videoMemoryReboot: 'Khởi động lại để thu hồi.',
        community: 'Cộng đồng',
        hostname: 'Tên máy chủ',
        hostnameUpdated: 'Đã cập nhật tên máy chủ. Khởi động lại để áp dụng.',
        ipType: {
          Wired: 'Có dây',
          Wireless: 'Không dây',
          Other: 'Khác'
        },
        hostnameInvalid:
          'Dùng chữ cái, chữ số và dấu gạch nối, tối đa 63 ký tự mỗi phần ngăn cách bởi dấu chấm. Không có dấu gạch nối ở đầu hoặc cuối phần.',
        hostnameFailed: 'Không thể đổi tên máy chủ',
        editHostname: 'Sửa tên máy chủ',
        docs: 'Tài liệu',
        hardware: 'Phần cứng',
        hardwareFaq: 'Câu hỏi thường gặp về phần cứng',
        disclaimer:
          'IronKVM: firmware cộng đồng được gia cố cho Sipeed NanoKVM. Không liên kết với Sipeed.',
        basedOn: 'dựa trên NanoKVM {{version}}'
      },
      appearance: {
        title: 'Giao diện',
        thisBrowser: 'Trình duyệt này',
        thisBrowserDesc: 'Chỉ lưu trong trình duyệt này. Các trình duyệt khác có cài đặt riêng.',
        deviceWide: 'Thiết bị',
        deviceWideDesc: 'Lưu trên thiết bị. Áp dụng cho mọi người mở thiết bị.',
        language: 'Ngôn ngữ',
        languageDesc: 'Chọn ngôn ngữ cho giao diện',
        webTitle: 'Tiêu đề trang web',
        webTitleDesc: 'Tùy chỉnh tiêu đề trang web',
        menuBar: {
          title: 'Thanh menu',
          mode: 'Chế độ hiển thị',
          modeDesc: 'Hiển thị thanh menu trên màn hình',
          modeOff: 'Tắt',
          modeAuto: 'Tự động ẩn',
          modeAlways: 'Luôn hiển thị',
          keyboardLedStatus: 'Chỉ báo khóa bàn phím',
          keyboardLedStatusDesc:
            'Hiển thị trạng thái Num Lock, Caps Lock và Scroll Lock của máy tính từ xa',
          icons: 'Biểu tượng menu con',
          iconsDesc: 'Hiển thị biểu tượng menu con trên thanh menu'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Trạng thái khóa bàn phím từ xa',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Bật',
        off: 'Tắt',
        unknown: 'Không rõ'
      },
      device: {
        title: 'Thiết bị',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'Độ sáng OLED',
          brightnessDescription: 'Mức thấp hơn giúp màn hình bền hơn',
          brightnessLevels: {
            '64': 'Thấp nhất',
            '96': 'Thấp',
            '128': 'Trung bình',
            '160': 'Cao',
            '207': 'Mặc định',
            '255': 'Tối đa'
          },
          0: 'Không bao giờ',
          15: '15 giây',
          30: '30 giây',
          60: '1 phút',
          180: '3 phút',
          300: '5 phút',
          600: '10 phút',
          1800: '30 phút',
          3600: '1 giờ'
        },
        advanced: 'Cài đặt nâng cao',
        cpuFreq: {
          title: 'Tần số CPU',
          description: 'Đặt xung nhịp CPU áp dụng ở lần khởi động tiếp theo',
          tip: 'CPU khởi động ở 850 MHz và được định mức cho 1000 MHz. Giá trị mới được áp dụng ở lần khởi động tiếp theo, không phải khi hệ thống đang chạy. 1000 MHz nằm trong thông số; nhiệt độ vẫn thấp hơn nhiều so với giới hạn ở cả hai mức.',
          running: 'Đang chạy: {{mhz}} MHz',
          rebootToApply: 'khởi động lại để áp dụng',
          rebootConfirm: 'Khởi động lại ngay để áp dụng {{mhz}} MHz?'
        },
        swap: {
          title: 'Hoán đổi',
          disable: 'Tắt',
          description: 'Đặt kích thước tệp hoán đổi',
          tip: 'Kích hoạt tính năng này có thể rút ngắn thời gian sử dụng thẻ SD của bạn!'
        },
        zram: {
          title: 'Hoán đổi nén (zram)',
          description: 'Hoán đổi trong RAM nén, thay vì trên thẻ SD',
          tip: 'zram giữ vùng hoán đổi ngoài thẻ SD, nên không gây hao mòn. Không có vùng hoán đổi trên đĩa phía sau: nếu zram đầy, kernel sẽ dừng một tiến trình thay vì phân trang chậm. Giới hạn bộ nhớ quy định lượng RAM tối đa zram có thể dùng.',
          unavailable: 'Các module kernel chưa được cài trên thiết bị này',
          inactive: 'Đã bật, nhưng thiết bị không khởi động',
          active: 'Đang hoạt động - {{used}} / {{total}}, {{ratio}}x',
          off: 'Tắt',
          detail: {
            algorithm: 'Thuật toán: {{algorithm}}',
            memory: 'Bộ nhớ đã dùng: {{used}} / {{limit}}',
            memoryNoLimit: 'Bộ nhớ đã dùng: {{used}}, không đặt giới hạn',
            counters:
              'Trang hoán đổi vào {{in}}, ra {{out}} (mọi thiết bị hoán đổi, từ lúc khởi động)'
          }
        },
        mouseJiggler: {
          title: 'Máy lắc lư chuột',
          description: 'Ngăn máy chủ từ xa ngủ',
          disable: 'Tắt',
          absolute: 'Chế độ tuyệt đối',
          relative: 'Chế độ tương đối'
        },
        mdns: {
          description: 'Kích hoạt dịch vụ khám phá mDNS',
          tip: 'Tắt đi nếu không cần thiết'
        },
        hdmi: {
          description: 'Kích hoạt HDMI/đầu ra màn hình',
          idleTimeoutTitle: 'Thời gian chờ khi không hoạt động',
          idleTimeoutDescription:
            'Dừng việc ghi hình HDMI sau khi không có người xem hoạt động trong',
          minutes: 'phút'
        },
        hidOnly: 'HID-Chế độ chỉ',
        hidOnlyDesc: 'Dừng mô phỏng các thiết bị ảo, chỉ giữ lại điều khiển HID cơ bản',
        disk: 'Đĩa ảo',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Mạng ảo',
        networkDesc: 'Gắn card mạng ảo trên máy chủ từ xa',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Máy chủ:',
          description:
            'Liên kết mạng riêng với máy chủ từ xa qua cáp USB. Máy chủ nhận một địa chỉ không có gateway và không có DNS, nên không thể truy cập mạng LAN của bạn thông qua IronKVM.',
          off: 'Tắt',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (cho máy chủ không hỗ trợ NCM)',
          rndis: 'RNDIS (không còn được cung cấp)',
          rndisNote: 'Liên kết này dùng RNDIS, không còn được cung cấp. Hãy chọn NCM hoặc ECM.',
          subnet: 'Mạng con',
          subnetDesc:
            'Một mạng IPv4 riêng, từ /24 đến /30. IronKVM dùng địa chỉ đầu tiên, máy chủ dùng địa chỉ thứ hai.',
          invalidSubnet: 'Nhập một mạng con, ví dụ 172.31.255.0/30.',
          apply: 'Áp dụng',
          confirm: 'Kết nối lại thiết bị USB?',
          reenumerate:
            'Áp dụng sẽ dựng lại kết nối USB. Máy chủ mất bàn phím, chuột và ổ đĩa ảo trong vài giây.'
        },
        audio: 'Loa ảo',
        audioDesc:
          'Cung cấp một card âm thanh USB cho máy chủ từ xa, để bạn có thể nghe âm thanh của nó. Máy chủ phải chọn nó làm thiết bị đầu ra. Thay đổi mục này sẽ dựng lại kết nối USB.',
        audioNote: 'Âm thanh có trong cả hai chế độ H.264 (WebRTC và Direct), không có trong MJPEG',
        console: 'Console nối tiếp',
        consoleDesc:
          'Cung cấp một cổng nối tiếp USB cho máy chủ từ xa, để đăng nhập vào IronKVM này khi không truy cập được mạng',
        consoleTip:
          'Bất kỳ ai điều khiển máy chủ từ xa đều nhận được lời nhắc đăng nhập vào IronKVM này. Đặt mật khẩu mạnh trước khi bật (Tài khoản - Đổi mật khẩu).',
        endpoints: {
          title: 'Endpoint USB',
          used: 'Đã dùng {{used}} / {{total}}',
          cost: 'dùng {{cost}}',
          needs: 'cần {{cost}}',
          full: 'Không đủ endpoint USB. Hãy tắt một mục khác trước.',
          inactive:
            'Đã bật nhưng không chạy: bộ điều khiển USB đã hết endpoint. Tắt một thiết bị khác và thiết bị này sẽ khởi động ngay.',
          explain:
            'Bộ điều khiển USB có một số lượng endpoint đầu vào cố định, và đây là số đếm của chúng. Nếu số thiết bị được bật vượt quá sức chứa, bàn phím và chuột được giữ lại, phần còn lại bị tắt.',
          error: 'Không kết nối được thiết bị. Hãy thử lại.',
          fitTogether: 'Có thể dùng cùng nhau: {{sets}}'
        },
        reboot: 'Khởi động lại',
        rebootDesc: 'Bạn có chắc chắn muốn khởi động lại IronKVM không?',
        okBtn: 'Có',
        cancelBtn: 'Không',
        rebootFailed: 'Khởi động lại thất bại'
      },
      network: {
        title: 'Mạng',
        wifi: {
          disconnectBtn: 'Ngắt kết nối',
          disconnectWarning:
            'Nếu bạn truy cập IronKVM qua mạng Wi-Fi này, trang này sẽ mất kết nối.',
          disconnected: 'Đã ngắt Wi-Fi',
          title: 'Wi-Fi',
          description: 'Cấu hình Wi-Fi',
          apMode: 'Chế độ AP đang bật, hãy kết nối Wi-Fi bằng cách quét mã QR',
          connect: 'Kết nối Wi-Fi',
          connectDesc1: 'Vui lòng nhập SSID mạng và mật khẩu',
          connectDesc2: 'Vui lòng nhập mật khẩu để tham gia mạng này',
          disconnect: 'Bạn có chắc muốn ngắt kết nối mạng không?',
          failed: 'Kết nối thất bại, vui lòng thử lại.',
          ssid: 'Tên',
          password: 'Mật khẩu',
          joinBtn: 'Tham gia',
          confirmBtn: 'OK',
          cancelBtn: 'Hủy'
        },
        tls: {
          description: 'Bật giao thức HTTPS',
          tip: 'Lưu ý: Sử dụng HTTPS có thể tăng độ trễ, đặc biệt trong chế độ video MJPEG.',
          restarting: 'Đang khởi động lại máy chủ của thiết bị, mất khoảng hai phút...',
          waiting: 'Đang chờ thiết bị phản hồi lại...',
          waitingHttp: 'Đang chuyển về http. Tải lại trang này nếu nó không tự mở.',
          failed: 'Không thể thay đổi cài đặt HTTPS',
          enableConfirm: 'Bật HTTPS?',
          disableConfirm: 'Tắt HTTPS?',
          confirmDesc:
            'Thao tác này đăng xuất bạn và khởi động lại máy chủ của thiết bị, mất khoảng hai phút. Sau đó trang sẽ mở {{url}}.',
          confirmOk: 'Tiếp tục',
          confirmCancel: 'Hủy'
        },
        ethernet: {
          title: 'Địa chỉ IP',
          description: 'Cấu hình cách IronKVM nhận địa chỉ trên mạng có dây',
          dhcp: 'DHCP',
          manual: 'Thủ công',
          networkDetails: 'Chi tiết mạng',
          interface: 'Giao diện',
          ipAddress: 'Địa chỉ IP',
          subnetMask: 'Mặt nạ mạng con',
          router: 'Bộ định tuyến',
          save: 'Áp dụng',
          invalidAddress: 'Vui lòng nhập địa chỉ IP hợp lệ',
          invalidMask: 'Vui lòng nhập mặt nạ mạng con hợp lệ, ví dụ 255.255.255.0 hoặc 24',
          invalidRouter: 'Vui lòng nhập địa chỉ bộ định tuyến hợp lệ',
          addressRequired: 'Cần có địa chỉ IP',
          maskRequired: 'Cần có mặt nạ mạng con',
          applyTitle: 'Thay đổi địa chỉ của IronKVM?',
          applyWarning:
            'Kết nối tới trang này sẽ mất. IronKVM áp dụng địa chỉ mới và chờ {{seconds}} giây để bạn kết nối tới nó tại địa chỉ đó. Kết nối được sẽ giữ lại thay đổi. Nếu không có gì kết nối tới nó, IronKVM khôi phục các cài đặt trước đó.',
          applyConfirm: 'Áp dụng',
          applyCancel: 'Hủy',
          applyFailed: 'Không áp dụng được địa chỉ',
          trialTitle: 'Đang chờ xác nhận',
          trialDhcp: 'IronKVM đang yêu cầu địa chỉ qua DHCP.',
          trialStatic: 'IronKVM hiện ở {{address}}.',
          trialInstruction:
            'Mở IronKVM tại địa chỉ mới và đăng nhập nếu được yêu cầu. Kết nối tới nó ở đó sẽ giữ lại thay đổi. Nếu không có gì kết nối tới IronKVM trong {{seconds}} giây, nó khôi phục các cài đặt trước đó.',
          trialOpen: 'Mở địa chỉ mới',
          trialKeep: 'Giữ các cài đặt này',
          trialKept: 'Địa chỉ mới đã được lưu',
          trialKeepFailed: 'Không giữ được cài đặt',
          trialGone: 'Thay đổi đã được khôi phục. Vui lòng thử lại.',
          unsaved: 'Thay đổi chưa lưu'
        },
        dns: {
          title: 'DNS',
          description: 'Cấu hình máy chủ DNS cho IronKVM',
          mode: 'Chế độ',
          dhcp: 'DHCP',
          manual: 'Thủ công',
          add: 'Thêm DNS',
          save: 'Lưu',
          invalid: 'Vui lòng nhập địa chỉ IP hợp lệ',
          noDhcp: 'Hiện không có DNS DHCP khả dụng',
          saved: 'Đã lưu cài đặt DNS',
          saveFailed: 'Không thể lưu cài đặt DNS',
          unsaved: 'Thay đổi chưa lưu',
          maxServers: 'Cho phép tối đa {{count}} máy chủ DNS',
          dnsServers: 'Máy chủ DNS',
          dhcpServersDescription: 'Máy chủ DNS được tự động lấy từ DHCP',
          manualServersDescription: 'Có thể chỉnh sửa máy chủ DNS thủ công',
          networkDetails: 'Chi tiết mạng',
          interface: 'Giao diện',
          ipAddress: 'Địa chỉ IP',
          subnetMask: 'Mặt nạ mạng con',
          router: 'Bộ định tuyến',
          none: 'Không có'
        }
      },
      vpn: {
        connect: 'Kết nối',
        connectDesc: 'Tham gia mạng {{name}}. Tắt sẽ ngắt kết nối mà không dừng dịch vụ.',
        kvmUrl: 'Địa chỉ KVM',
        moreTip: 'Thao tác khác',
        restartTip: 'Khởi động lại',
        stopTip: 'Dừng',
        updateTip: 'Cập nhật lên {{version}}',
        loading: 'Đang tải...',
        okBtn: 'Có',
        cancelBtn: 'Không',
        restart: 'Khởi động lại {{name}}?',
        stop: 'Dừng {{name}}?',
        stopDesc:
          'Daemon sẽ dừng ngay. Khởi động cùng hệ thống là một công tắc riêng và vẫn giữ nguyên.',
        update: 'Cập nhật {{name}} lên {{version}}?',
        updateDesc: 'Daemon sẽ khởi động lại nếu đang chạy. Thông tin đăng nhập được giữ nguyên.',
        notInstall: '{{name}} chưa được cài đặt.',
        install: 'Cài đặt',
        installing: 'Đang cài đặt',
        installFailed: 'Cài đặt thất bại',
        retry: 'Thử lại',
        notRunning: '{{name}} không chạy. Hãy khởi động nó để tiếp tục.',
        run: 'Khởi động',
        boot: 'Khởi động cùng hệ thống',
        bootDesc: 'Khởi động {{name}} khi KVM khởi động.',
        control: 'Máy chủ điều khiển',
        connected: 'Đã kết nối',
        disconnected: 'Chưa kết nối',
        deviceName: 'Tên thiết bị',
        deviceIP: 'IP thiết bị',
        account: 'Tài khoản',
        version: 'Phiên bản',
        uptime: 'Thời gian hoạt động',
        peers: 'Peer',
        noPeers: 'Chưa có peer nào.',
        online: 'Trực tuyến',
        offline: 'Ngoại tuyến',
        memory: 'Bộ nhớ',
        daemonRss: 'Daemon',
        group: 'Nhóm tiện ích bổ sung',
        high: 'bị điều tiết khi vượt {{size}}',
        max: 'bị kernel dừng khi vượt {{size}}',
        noGroup: 'Bo mạch này không có nhóm bộ nhớ cho tiện ích bổ sung.',
        uninstall: 'Gỡ cài đặt {{name}}',
        uninstallDesc:
          'Bạn có chắc chắn muốn gỡ cài đặt {{name}} không? Thông tin đăng nhập vẫn được giữ trên bo mạch.',
        blocked:
          '{{other}} đang chạy hoặc khởi động cùng hệ thống. Mỗi lúc chỉ chạy được một VPN: hãy dừng {{other}} và tắt khởi động cùng hệ thống của nó trước.',
        swap: {
          title: 'Bộ nhớ hoán đổi',
          tip: 'Nếu daemon thiếu bộ nhớ, hãy thử bật bộ nhớ hoán đổi. Thao tác này đặt kích thước tệp hoán đổi mặc định là 256MB, có thể điều chỉnh trong "Cài đặt > Thiết bị".',
          failed: 'Không thể thay đổi bộ nhớ swap'
        },
        copy: 'Sao chép',
        copied: 'Đã sao chép liên kết',
        copyFailed: 'Không thể sao chép liên kết. Hãy chọn và sao chép thủ công.',
        open: 'Mở',
        checkAgain: 'Kiểm tra lại',
        notSignedIn: 'Chưa đăng nhập. Hãy hoàn tất đăng nhập qua liên kết rồi kiểm tra lại.',
        checkFailed: 'Không thể kiểm tra trạng thái đăng nhập',
        loginWaiting: 'Trang này kiểm tra vài giây một lần và sẽ tiếp tục khi bạn đã đăng nhập.',
        uninstallFailed: 'Gỡ cài đặt thất bại',
        loginFailed: 'Đăng nhập thất bại'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Tải xuống',
        package: 'gói cài đặt',
        unzip: 'và giải nén nó',
        notLogin:
          'Thiết bị chưa được liên kết. Vui lòng đăng nhập và liên kết thiết bị này với tài khoản của bạn.',
        urlPeriod: 'URL này có hiệu lực trong 10 phút',
        login: 'Đăng nhập',
        logout: 'Đăng xuất',
        logoutDesc: 'Bạn có chắc chắn muốn đăng xuất không?',
        manualIntro: 'Hoặc cài đặt thủ công qua SSH:',
        copyBinaries: 'Sao chép tailscale và tailscaled vào {{dir}} trên IronKVM',
        linksFile: 'Trong cùng thư mục, tạo một tệp tên links chứa hai dòng sau:',
        rebootRefresh: 'Khởi động lại IronKVM, rồi làm mới trang này'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Thiết bị này chưa tham gia mạng NetBird nào. Hãy tham gia bằng setup key, hoặc đăng nhập bằng SSO.',
        setupKey: 'Khóa thiết lập',
        setupKeyPlaceholder: 'Dán setup key từ bảng điều khiển NetBird',
        join: 'Tham gia',
        or: 'hoặc',
        sso: 'Đăng nhập bằng SSO',
        urlPeriod: 'URL này có hiệu lực trong 10 phút',
        logout: 'Hủy đăng ký',
        logoutDesc:
          'Hủy đăng ký sẽ xóa peer này khỏi tài khoản NetBird của bạn và xóa cấu hình của nó tại đây. Để tham gia lại cần setup key hoặc đăng nhập SSO, và peer có thể nhận IP mới. Tiếp tục?',
        joinFailed: 'Không thể tham gia mạng'
      },
      update: {
        title: 'Kiểm tra cập nhật',
        queryFailed: 'Lấy phiên bản thất bại',
        updateFailed: 'Cập nhật thất bại. Vui lòng thử lại.',
        isLatest: 'Bạn đã có phiên bản mới nhất.',
        available: 'Có bản cập nhật mới. Bạn có chắc chắn muốn cập nhật không?',
        updating: 'Bắt đầu cập nhật. Vui lòng chờ...',
        confirm: 'Xác nhận',
        cancel: 'Hủy',
        preview: 'Bản cập nhật xem trước',
        previewDesc: 'Nhận quyền truy cập sớm vào các tính năng và cải tiến mới',
        previewTip:
          'Xin lưu ý rằng các bản phát hành xem trước có thể có lỗi hoặc chức năng chưa hoàn chỉnh!',
        customServer: {
          title: 'Máy chủ cập nhật tùy chỉnh',
          desc: 'Kiểm tra và tải xuống các bản cập nhật trực tuyến từ máy chủ được chỉ định',
          invalidUrl:
            'Nhập thư mục máy chủ HTTP hoặc HTTPS hợp lệ, không chứa truy vấn, phân đoạn hoặc latest.json.',
          loadFailed: 'Không thể tải cấu hình máy chủ cập nhật.',
          saveFailed: 'Không thể lưu cấu hình máy chủ cập nhật.',
          saved: 'Đã lưu cấu hình máy chủ cập nhật.',
          save: 'Lưu',
          confirmTitle: 'Sử dụng máy chủ cập nhật tùy chỉnh?',
          confirmDesc:
            'SHA-512 chỉ kiểm tra xem gói có khớp với tệp kê khai do máy chủ này cung cấp hay không. Điều này không chứng minh rằng gói đó là bản phát hành IronKVM chính thức. Máy chủ bị lỗi hoặc độc hại có thể khiến thiết bị không thể sử dụng, gây mất dữ liệu hoặc xâm phạm hệ thống.',
          confirm: 'Vẫn sử dụng',
          useSipeed: 'Dùng máy chủ chính thức của Sipeed',
          previewDisabled:
            'Không thể sử dụng Bản cập nhật xem trước khi máy chủ cập nhật tùy chỉnh đang được bật.'
        },
        offline: {
          chooseFile: 'Chọn tệp',
          installing: 'Tải lên xong. Đang cài đặt...',
          noFile: 'Chưa chọn tệp',
          title: 'Cập nhật ngoại tuyến',
          desc: 'Cập nhật thông qua gói cài đặt cục bộ',
          upload: 'Tải lên',
          checksumPlaceholder: 'Tổng kiểm SHA-256 (không bắt buộc)',
          invalidChecksum: 'Tổng kiểm SHA-256 phải chứa 64 ký tự thập lục phân.',
          checksumMismatch: 'Xác minh SHA-256 không thành công. Gói có thể đã bị hỏng.',
          invalidName:
            'Định dạng tên tệp không hợp lệ. Vui lòng tải xuống từ bản phát hành GitHub.',
          updateFailed: 'Cập nhật thất bại. Vui lòng thử lại.'
        },
        updateTo: 'Cập nhật lên {{version}}',
        updateConfirmDesc:
          'Thiết bị cài đặt bản cập nhật và khởi động lại máy chủ. Trang này sẽ tải lại khi máy chủ hoạt động trở lại.',
        releaseNotes: 'Ghi chú phát hành'
      },
      account: {
        title: 'Tài khoản',
        webAccount: 'Tên tài khoản web',
        role: 'Vai trò',
        roles: { admin: 'Quản trị viên', user: 'Người dùng' },
        password: 'Mật khẩu',
        updateBtn: 'Update',
        logoutBtn: 'Đăng xuất',
        logoutDesc: 'Bạn có chắc chắn muốn đăng xuất không?',
        okBtn: 'Có',
        cancelBtn: 'Không',
        users: {
          title: 'Người dùng',
          create: 'Tạo người dùng',
          enabled: 'Đã bật',
          disabled: 'Đã tắt',
          deviceOwner: 'Chủ sở hữu thiết bị',
          resetPassword: 'Đặt lại mật khẩu',
          delete: 'Xóa',
          deleteConfirm: 'Xóa người dùng này và thu hồi mọi phiên của họ?',
          created: 'Đã tạo người dùng',
          deleted: 'Đã xóa người dùng',
          passwordUpdated: 'Đã cập nhật mật khẩu',
          loadFailed: 'Không tải được danh sách người dùng',
          saveFailed: 'Không lưu được người dùng',
          deleteFailed: 'Không xóa được người dùng'
        }
      },
      apiKeys: {
        mcpNote: 'Các khóa này không dùng được cho MCP, vì MCP có khóa riêng trên trang MCP.',
        metricsUrl: 'URL số liệu',
        monitoring: 'Giám sát',
        monitoringDesc:
          'Prometheus đọc số liệu bằng một khóa API từ trang này, gửi dưới dạng token Bearer. Mọi vai trò đều đọc được.',
        scrapeConfig: 'Cấu hình scrape cho Prometheus',
        title: 'Khóa API',
        description:
          'Khóa hoạt động với tư cách chủ sở hữu, với vai trò của người dùng đó. Gửi nó dưới dạng Authorization: Bearer <key> cho metrics và API, hoặc dưới dạng X-Auth-Token cho Redfish.',
        name: 'Tên',
        namePlaceholder: 'Mục đích của khóa, ví dụ prometheus',
        nameRequired: 'Hãy đặt tên cho khóa',
        nameTooLong: 'Tên dài tối đa 64 ký tự',
        unnamed: '(không tên)',
        create: 'Tạo khóa',
        created: 'Tạo lúc',
        owner: 'Chủ sở hữu',
        empty: 'Không có khóa API',
        newKeyTitle: 'Khóa API mới của bạn',
        newKeyWarning:
          'Hãy sao chép khóa ngay. Khóa không được lưu và không thể hiển thị lại. Nếu làm mất, hãy thu hồi và tạo khóa khác.',
        copy: 'Sao chép',
        copied: 'Đã sao chép',
        copyFailed: 'Sao chép thất bại. Vui lòng sao chép thủ công.',
        done: 'Xong',
        revoke: 'Thu hồi',
        revokeConfirmTitle: 'Thu hồi khóa API này?',
        revokeConfirmDesc: 'Mọi thứ đang dùng "{{name}}" sẽ ngừng hoạt động ngay lập tức.',
        revoked: 'Đã thu hồi khóa API',
        loadFailed: 'Không tải được khóa API',
        createFailed: 'Không tạo được khóa API',
        revokeFailed: 'Không thu hồi được khóa API',
        cancelBtn: 'Hủy'
      }
    },
    picoclaw: {
      title: 'PicoClaw Trợ lý',
      empty: 'Mở bảng điều khiển và bắt đầu một nhiệm vụ.',
      inputPlaceholder: 'Mô tả những gì bạn muốn PicoClaw làm',
      newConversation: 'Cuộc trò chuyện mới',
      processing: 'Đang xử lý...',
      agent: {
        defaultTitle: 'Trợ lý chung',
        defaultDescription: 'Trợ giúp chung về trò chuyện, tìm kiếm và không gian làm việc.',
        kvmTitle: 'Điều khiển từ xa',
        kvmDescription: 'Vận hành máy chủ từ xa thông qua IronKVM.',
        switched: 'Vai trò đại lý đã chuyển đổi',
        switchFailed: 'Chuyển đổi vai trò đại lý không thành công'
      },
      send: 'Gửi',
      cancel: 'Hủy',
      status: {
        connecting: 'Đang kết nối với cổng...',
        connected: 'Phiên PicoClaw đã kết nối',
        disconnected: 'Phiên PicoClaw đã ngắt kết nối',
        stopped: 'Đã gửi yêu cầu dừng',
        runtimeStarted: 'Runtime PicoClaw đã bắt đầu',
        runtimeStartFailed: 'Không khởi động được runtime PicoClaw',
        runtimeStopped: 'Runtime PicoClaw đã dừng',
        runtimeStopFailed: 'Không dừng được runtime PicoClaw',
        controlSwitchedToMCP: 'Quyền điều khiển đã chuyển sang dịch vụ MCP bên ngoài'
      },
      connection: {
        runtime: {
          checking: 'Đang kiểm tra',
          restoring: 'Đang khôi phục PicoClaw',
          ready: 'Runtime đã sẵn sàng',
          stopped: 'Đã dừng runtime',
          blockedByMCP: 'Điều khiển MCP bên ngoài đang hoạt động',
          readyBlockedByMCP:
            'Runtime đang chạy, nhưng MCP bên ngoài hiện đang điều khiển đầu vào thiết bị.',
          readyWithoutControl:
            'Runtime đang chạy. Hãy cấp quyền điều khiển thiết bị cho PicoClaw trước khi kết nối lại.',
          unavailable: 'Runtime không khả dụng',
          configError: 'Lỗi cấu hình'
        },
        transport: {
          connecting: 'Đang kết nối',
          connected: 'Đã kết nối',
          disconnected: 'Đã ngắt kết nối',
          reconnect: 'Kết nối lại',
          reconnectDescription: 'Kết nối lại với phiên PicoClaw đang chạy.',
          reconnectBlocked: 'PicoClaw cần quyền điều khiển thiết bị trước khi kết nối lại.'
        },
        run: {
          idle: 'Nhàn rỗi',
          busy: 'Bận'
        }
      },
      message: {
        toolAction: 'Hành động',
        observation: 'Quan sát',
        screenshot: 'Ảnh chụp màn hình'
      },
      overlay: {
        locked: 'PicoClaw đang điều khiển thiết bị. Việc nhập thủ công bị tạm dừng.'
      },
      control: {
        picoclaw: 'Điều khiển thiết bị: PicoClaw',
        picoclawDescription:
          'PicoClaw có thể gửi thao tác bàn phím và chuột. Thao tác thủ công có thể bị tạm dừng.',
        mcp: 'Điều khiển thiết bị: MCP bên ngoài',
        mcpDescription:
          'MCP bên ngoài có thể ghi vào thiết bị. PicoClaw sẽ không tiếp quản đầu vào.',
        off: 'Điều khiển thiết bị: tắt',
        offDescription:
          'AI sẽ không gửi thao tác bàn phím hay chuột. Điều khiển thủ công vẫn dùng được.',
        transitioning: 'Điều khiển thiết bị: đang chuyển',
        transitioningDescription: 'Đang đồng bộ quyền điều khiển thiết bị. Vui lòng chờ.',
        grant: 'Cấp quyền điều khiển',
        release: 'Nhả',
        releasing: 'Đang nhả...',
        switching: 'Đang chuyển...',
        releasingLabel: 'Điều khiển thiết bị: đang nhả',
        releasingDescription:
          'Đang trả lại quyền điều khiển thiết bị. PicoClaw đã dừng các thao tác ghi đang chạy.',
        granted: 'Đã cấp quyền điều khiển PicoClaw',
        released: 'Đã nhả quyền điều khiển PicoClaw',
        grantFailed: 'Không thể cấp quyền điều khiển PicoClaw',
        releaseFailed: 'Không thể nhả quyền điều khiển PicoClaw',
        grantConfirmTitle: 'Chuyển điều khiển thiết bị sang PicoClaw?',
        grantConfirmDesc: 'Các thao tác ghi thiết bị của MCP bên ngoài sẽ bị gián đoạn.'
      },
      install: {
        install: 'Cài đặt PicoClaw',
        installing: 'Đang cài đặt PicoClaw',
        success: 'PicoClaw cài đặt thành công',
        failed: 'Không cài đặt được PicoClaw',
        uninstalling: 'Đang gỡ cài đặt runtime...',
        uninstalled: 'Đã gỡ cài đặt thành công runtime.',
        uninstallFailed: 'Gỡ cài đặt không thành công.',
        requiredTitle: 'PicoClaw chưa được cài đặt',
        requiredDescription: 'Cài đặt PicoClaw trước khi bắt đầu runtime PicoClaw.',
        progressDescription: 'PicoClaw đang được tải xuống và cài đặt.',
        stages: {
          preparing: 'Đang chuẩn bị',
          downloading: 'Đang tải xuống',
          extracting: 'Đang giải nén',
          verifying: 'Đang xác minh',
          installing: 'Đang cài đặt',
          installed: 'Đã cài đặt',
          install_timeout: 'Đã hết thời gian',
          install_failed: 'Không thành công'
        }
      },
      model: {
        requiredTitle: 'Cần có cấu hình mô hình',
        requiredDescription:
          'Định cấu hình mô hình PicoClaw trước khi sử dụng trò chuyện PicoClaw.',
        docsTitle: 'Hướng dẫn cấu hình',
        docsDesc: 'Các mô hình và giao thức được hỗ trợ',
        menuLabel: 'Cấu hình mô hình',
        modelIdentifier: 'Định danh mô hình',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Khóa API',
        apiKeyPlaceholder: 'Nhập khóa API của mô hình',
        save: 'Lưu',
        saving: 'Đang lưu',
        saved: 'Đã lưu cấu hình mô hình',
        saveFailed: 'Không lưu được cấu hình mô hình',
        invalid: 'Bắt buộc nhập mã định danh mô hình, API Base URL và khóa API'
      },
      uninstall: {
        menuLabel: 'Gỡ cài đặt',
        confirmTitle: 'Gỡ cài đặt PicoClaw',
        confirmContent:
          'Bạn có chắc chắn muốn gỡ cài đặt PicoClaw không? Thao tác này sẽ xóa tệp thực thi và tất cả các tệp cấu hình.',
        confirmOk: 'Gỡ cài đặt',
        confirmCancel: 'Hủy'
      },
      history: {
        title: 'Lịch sử',
        loading: 'Đang tải phiên...',
        emptyTitle: 'Chưa có lịch sử',
        emptyDescription: 'Các phiên PicoClaw trước đó sẽ xuất hiện ở đây.',
        loadFailed: 'Không tải được lịch sử phiên',
        deleteFailed: 'Không xóa được phiên',
        deleteConfirmTitle: 'Xóa phiên',
        deleteConfirmContent: 'Bạn có chắc chắn muốn xóa "{{title}}" không?',
        deleteConfirmOk: 'Xóa',
        deleteConfirmCancel: 'Hủy',
        messageCount_other: '{{count}} tin nhắn',
        messageCount: '{{count}} tin nhắn'
      },
      config: {
        startRuntime: 'Bắt đầu PicoClaw',
        stopRuntime: 'Dừng PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Chuyển quyền điều khiển sang PicoClaw?',
        enableConfirmDesc: 'Khởi động PicoClaw sẽ tắt dịch vụ MCP bên ngoài.',
        enableConfirmOk: 'Bắt đầu PicoClaw',
        enableConfirmCancel: 'Hủy',
        title: 'Bắt đầu PicoClaw',
        description: 'Bắt đầu runtime để sử dụng trợ lý PicoClaw.',
        switchFromMCP: 'Chuyển sang PicoClaw và bắt đầu',
        takeoverAndStart: 'Tiếp quản và bắt đầu'
      }
    },
    error: {
      title: 'Chúng tôi đã gặp sự cố',
      refresh: 'Làm mới',
      panel: 'Phần này của trang đã ngừng hoạt động',
      retry: 'Thử lại'
    },
    fullscreen: {
      toggle: 'Chuyển đổi toàn màn hình'
    },
    input: {
      disconnected: 'Bàn phím và chuột chưa được kết nối',
      disconnectedTls:
        'Trình duyệt đã từ chối kết nối bảo mật truyền bàn phím và chuột, và nó làm vậy mà không hỏi. Chứng chỉ do thiết bị này tạo ra chưa được tin cậy. Hãy mở địa chỉ này trong tab mới, chấp nhận chứng chỉ, rồi tải lại. Cài đặt chứng chỉ là cách khắc phục đáng tin cậy.',
      disconnectedNever:
        'Không mở được kết nối truyền bàn phím và chuột. Phần còn lại của trang vẫn hoạt động vì không dùng kết nối này. Hãy kiểm tra xem có gì giữa bạn và thiết bị đang chặn nó không.',
      disconnectedDropped:
        'Kết nối truyền bàn phím và chuột đã bị mất và chưa khôi phục. Nó tự kết nối lại sau khi khởi động lại; nếu tình trạng này kéo dài, hãy tải lại trang.',
      hidDisabled: 'HID đã bị tắt trên thiết bị này (/boot/disable_hid).',
      keyFailed: 'Không thể gửi phím.'
    },
    speaker: { title: 'Loa', unmute: 'Bật tiếng', mute: 'Tắt tiếng' },
    upstream: {
      check: 'Kiểm tra cập nhật',
      updateTo: 'Cập nhật lên {{version}}',
      confirm: 'Cập nhật {{name}} lên {{version}}?',
      confirmDesc:
        'Bản phát hành mới được tải từ GitHub và kiểm tra bằng checksum mà nó công bố. Nếu có lỗi, phiên bản hiện tại vẫn được giữ.',
      ok: 'Cập nhật',
      upToDate: 'Đã là mới nhất',
      builtIn: 'tích hợp',
      checkFailed: 'Không thể kiểm tra cập nhật: {{error}}',
      unverifiable: 'Phiên bản {{version}} không được cung cấp: {{reason}}',
      inUse: 'Hiện không thể cập nhật: {{reason}}',
      running: 'Đang cập nhật lên {{version}}...',
      done: 'Đã cập nhật {{name}} lên {{version}}',
      failed: 'Lần cập nhật gần nhất thất bại: {{error}}'
    },
    menu: {
      collapse: 'Thu gọn Menu',
      expand: 'Mở rộng Menu',
      more: 'Thêm'
    },
    ion: {
      checking: 'Đang kiểm tra bộ nhớ video trước khi bắt đầu luồng...',
      warn: 'Bộ nhớ video sắp hết. Chỉ một lần khởi động lại máy chủ cũng sẽ làm cạn nó. Hãy khởi động lại khi thuận tiện.',
      criticalTitle: 'Không đủ bộ nhớ video để bắt đầu luồng',
      criticalBody:
        'Bắt đầu video sẽ làm cạn bộ nhớ dành riêng và dừng máy chủ. Mọi chức năng khác vẫn hoạt động, bao gồm điều khiển nguồn và khởi động lại. Chỉ khởi động lại IronKVM mới thu hồi được bộ nhớ này.',
      criticalContinue: 'Vẫn bắt đầu video',
      criticalReboot: 'Khởi động lại IronKVM',
      criticalRebooting: 'Đang khởi động lại...'
    }
  }
};

export default vi;
