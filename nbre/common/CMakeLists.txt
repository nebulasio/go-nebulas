add_library(nbre_common
  util/base58.cpp
  util/base64.cpp
  util/byte.cpp
  ipc/shm_queue.cpp
  ipc/shm_bookkeeper.cpp
  ipc/shm_session.cpp
  ipc/shm_service_construct_helper.cpp
  ipc/shm_service_op_queue.cpp
  ipc/shm_service_recv_handler.cpp
  ipc/shm_service.cpp
  ipc/shm_queue_watcher.cpp
  quitable_thread.cpp
  exception_queue.cpp
  ir_conf_reader.cpp
  configuration.cpp)

target_link_libraries(nbre_common glog boost_program_options)
