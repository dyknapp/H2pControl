# See
# https://github.com/nix-community/poetry2nix/blob/master/overrides/default.nix
pkgs: # nixpkgs
final: # final python package set
prev: # previous python package set
{ 

  # THIS IS FOR PYQT6 FOR ARTIQ 9+
  # https://pypi.org/project/PyQt6-Qt6/
  # https://github.com/nix-community/poetry2nix/blob/1fb01e90771f762655be7e0e805516cd7fa4d58e/overrides/default.nix#L2871
#  pyqt6-qt6 = prev.pyqt6-qt6.overrideAttrs (old: {
#    autoPatchelfIgnoreMissingDeps = [ "libmysqlclient.so.21" "libmimerapi.so" "libQt6*" ];
#    propagatedBuildInputs = old.propagatedBuildInputs or [ ] ++ [
#      pkgs.qt6.full # Isn't this kind of cheating? The whole point of pyqt6-qt6
#                    # is to provide only what pyqt6 needs, not the whole qt6.full.
#      pkgs.libxkbcommon
#      pkgs.gtk3
#      pkgs.speechd
#      pkgs.gst
#      pkgs.gst_all_1.gst-plugins-base
#      pkgs.gst_all_1.gstreamer
#      pkgs.postgresql.lib
#      pkgs.unixODBC
#      pkgs.pcsclite
#      pkgs.xorg.libxcb
#      pkgs.xorg.xcbutil
#      pkgs.xorg.xcbutilcursor
#      pkgs.xorg.xcbutilerrors
#      pkgs.xorg.xcbutilimage
#      pkgs.xorg.xcbutilkeysyms
#      pkgs.xorg.xcbutilrenderutil
#      pkgs.xorg.xcbutilwm
#      pkgs.libdrm
#      pkgs.pulseaudio
#    ];
#  });
#
#    # https://pypi.org/project/PyQt6/
#  pyqt6 = prev.pyqt6.overrideAttrs (old: {
#    buildInputs = old.buildInputs or [ ] ++ [
#      final.pyqt6-qt6
#    ];
#  });
#
  # THIS IS FOR PYQT5 FOR ARTIQ 8
  pyqt5-qt5 = prev.pyqt5-qt5.overrideAttrs (old: {
    autoPatchelfIgnoreMissingDeps = [ "libmysqlclient.so.21" "libmimerapi.so" "libQt5*" ];
    propagatedBuildInputs = (old.propagatedBuildInputs or [ ]) ++ [
      pkgs.qt5.full
      pkgs.libxkbcommon
      pkgs.gtk3
      pkgs.speechd
      pkgs.gst_all_1.gst-plugins-base
      pkgs.gst_all_1.gstreamer
      pkgs.postgresql.lib
      pkgs.unixODBC
      pkgs.pcsclite
      pkgs.xorg.libxcb
      pkgs.xorg.xcbutil
      pkgs.xorg.xcbutilcursor
      pkgs.xorg.xcbutilerrors
      pkgs.xorg.xcbutilimage
      pkgs.xorg.xcbutilkeysyms
      pkgs.xorg.xcbutilrenderutil
      pkgs.xorg.xcbutilwm
      pkgs.libdrm
      pkgs.pulseaudio
    ];
  });

  pyqt5 = prev.pyqt5.overrideAttrs (old: {
    buildInputs = (old.buildInputs or [ ]) ++ [ final.pyqt5-qt5 ];
  });

}
