add_library(nbre_common
  util/base58.cpp
  util/base64.cpp
  util/byte.cpp
  ir_conf_reader.cpp)

target_link_libraries(nbre_common glog)
