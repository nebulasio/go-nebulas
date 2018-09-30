add_library(nbre_common
  util/base58.cpp
  util/base64.cpp
  util/byte.cpp
  ipc/shm_queue.cpp
  ir_conf_reader.cpp
  configuration.cpp)

target_link_libraries(nbre_common glog boost_program_options)
