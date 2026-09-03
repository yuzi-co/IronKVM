list(APPEND ADD_REQUIREMENTS vision basic peripheral)
list(APPEND ADD_INCLUDE "include")

# The `vision` requirement above pulls in the whole opencv package, so the
# linker is handed every opencv module whether or not this component calls into
# it. Without --as-needed it writes one NEEDED entry for each, and the device
# loader then maps libraries that nothing here uses: libopencv_video.so.409
# alone drags in libopencv_dnn, libopencv_calib3d, libopencv_features2d and
# libopencv_flann, about 6.1MB that no symbol in libkvm.so reaches.
#
# --as-needed records a dependency only when it resolves an undefined symbol,
# so the four opencv modules this component does use survive and the rest do
# not. The committed library was trimmed by hand with
# `patchelf --remove-needed` before this flag existed; the flag is what makes
# the next rebuild reproduce that state instead of undoing it.
list(APPEND ADD_LINK_DEFINITIONS_PRIVATE -Wl,--as-needed)

append_srcs_dir(ADD_SRCS "src")

register_component(DYNAMIC)
