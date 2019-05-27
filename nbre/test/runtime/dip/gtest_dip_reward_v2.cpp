// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "common/common.h"
#include "runtime/dip/dip_reward_v2.h"
#include "runtime/util.h"
#include "test/runtime/util.h"
#include <gtest/gtest.h>

TEST(test_runtime_dip_reward_v2, dip_votes_float) {
  std::map<neb::address_t, std::map<neb::address_t, neb::floatxx_t>>
      acc_contract_votes;

  std::vector<std::string> acc_contract_votes_str(
      {"n1UVddrGJNQr78tdTaYSxypqW7MnAL9uReL,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1111383670",
       "n1ZzXujRgwTb5U99bigt29P4RZ2MX82nSfe,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1PWP2cumPpNMK1MdqTdPp1SmxxRwnRY7A2,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1QvKS22AudZjPpsGyaqaeWKyUbeasHRTUW,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1FG5DPnoozH83vtcahXbibg8JKfo2SbE2o,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1153726883",
       "n1UdNL5ubPCFRVSyxp9DHcocYuc5oytX7YK,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1QaKyNo8meuBP6jSQjx8oRQCbSq5nS8HxV,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1141795420",
       "n1M8f5zSc3GXo3n6DFfF6ypMqJe3TxHDfNr,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1047306328",
       "n1HbtAJfVexj4YoJ7eiSGRynq96s5mZrv2M,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "667511789",
       "n1JsdNK15j12woqrQiaRYj7vYmgX41QzTFy,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1217888379",
       "n1RzkrBeTh1uMFZQGx575bgd1d1bsZk6mNo,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1092214512",
       "n1XBGikvoTU6huJcJHkP8woian8pvaRfHQk,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1092285986",
       "n1YpsSQugYFNXtfLQ7p2ckg8h8YYxbmfNVC,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1096741485",
       "n1MKEcM9F6BukDSV2P2GiFfmoFfszScUcji,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1116520450",
       "n1WPBRtz949TjRkHpYiCb5PECdfGKuCU94B,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1023854480",
       "n1ah1RWT72WGWe35HmH5kMUnqihsCJ4CVZv,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1VhRq6WzJUXoStdKeD2wYsv6mqhYzTnSp6,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1029422475",
       "n1WE6mypAzspXzTA6FLasDNeoMXSkmM56nu,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1075413143",
       "n1FzjsYNvqYtEHeijjHg3hDyua8g61SdUTB,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1UNvCeRoM3N9JWye6AoeHoi9VYdCKcTpsb,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1007521842",
       "n1XuzAnZd51nMMdTZWC5vJZgEXvWWyuzAPU,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1021354700",
       "n1RmGMscmzA67HeiqzkZsc46LABN7S16923,"
       "n22f8VKJ3ymzpAH4iZWJLUEXMBCTv6cbSF9,"
       "1120750436",
       "n1d5QiPS2TytTSCaYZAN8NbhTYuLpTEWc7C,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1QnPJN9kydNrM6CVVGymDPSC1Hy4dYs6tG,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "940692947",
       "n1cZF8Kp4o8PmSyhPvmkn4VEp3phUXP1Qke,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1035609660",
       "n1GTXax33p4qkk5HVpmcs8F7gVZ1ngYmiFr,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1131735736",
       "n1YFU5oWhDDaDJx9EknAiyQh4PUfVkcnvRs,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1167363398",
       "n1YwT4dvDwkiFiLDpQU17CqphTEvZgz1s8F,"
       "n1fPbCgC2Gt4ABcxdeBVbYPv6JHNVJJ2G6q,"
       "1173429509",
       "n1KtQGtCHupxMj3FZYXY6WDPUFGMWnxUnbs,"
       "n1fPbCgC2Gt4ABcxdeBVbYPv6JHNVJJ2G6q,"
       "1178352459",
       "n1PP9Q78G1vYmdeRWh9vGLEx9bTaAgqSfXk,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1T5pQRZAAq7SmfD8zApzRbmitS4MhCXLRr,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1102357606",
       "n1S7dHrRwNXNy4SUg2tyP5zjaMh1PkzwVnv,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1031904752",
       "n1NgiTP6PgDC9xMf36V4rE6ebMRh34EJQj9,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1200058367",
       "n1V9Rzcvc2pqerWvaDNT5jKXpawGZuUBjd8,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1135753150",
       "n1J7RXfJBgjZcBwJxhbD7m3hyHh7UhGDBQF,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1bTzegyXNAgXvWbMfovTDfeadPgxgjftmv,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1076672227",
       "n1QqmK4A6XS1AsVe7WH1UVdQ8KVudjMJVrS,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1GxC1TMUMyJv14tC3UBm3eKf3aNpetZZ6R,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1184156643",
       "n1X6Gq1J9G5V9UEtaKnB4XmWMRxa7d1vN4x,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1135350325",
       "n1NXTt5CRyBXVHA2D7VHpS6RmiqdpT9XGTG,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "0",
       "n1QCv4SHv8o9yHBAYBKNJUfsdoMYRhhh2HY,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1205250502",
       "n1Sq9HnsvN9YspD9qd3mdGU8ApBwpnF5Q9X,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1178267371",
       "n1Q9rfsKWrVCwPiXtQqzrmFvzDz8te4tfMV,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "934161877",
       "n1Gq5HcPvxbVb8e3TA8HNUZjwyKKKfuRVE8,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1WRDTpzaQnAUiPwiXJqXGvoB2uE2SWj4Tz,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1035968278",
       "n1Un9gVZaFpA9Tm2qc34xvq4S5bcUJzUcoG,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1VhLd3h3MYBDSRyZ54H9zdVb4v6L2Jptw1,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "0",
       "n1dBZ4yXoMLp4x8snpXojw9dWf5iBmaX2D9,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1090870797",
       "n1ZG8DzNqc6apPWPVJ8XG76nJx4aMjwbYoB,"
       "n1y395hWWXVUwz7dWG7bjjTrurkfDgNbhhn,"
       "0",
       "n1dqCLcSJoEQ2GayZERpQHZBKbTbnrrgKWq,"
       "n22f8VKJ3ymzpAH4iZWJLUEXMBCTv6cbSF9,"
       "1151933423",
       "n1NmKhroXEMez1HH6q73jqNkaTYoqS1zLHo,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1023885422",
       "n1K7rJP1ACQTvwauthPPg9o6iEzMSpT5Ti9,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "0",
       "n1UdzmdnypdovCN4UoHWtnAANGU7uNSzfgT,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1189322507",
       "n1WtqfkxjAuxWJawKJk4QqtBPZdbX6XrGYr,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1WhDVwHWf6Jdt1tBZ1ykdwRaiaxEpgkYBR,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1072785069",
       "n1TwsZVWXScFfp48gHWuuvFN47KXBNM2j7t,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1QUMEs7qkvCDuimT2zYsbf62z9pRtofQiP,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1XMw7d7se737dogVYiBHckj7kNFJ51dBMs,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1027952333",
       "n1W3BHxSRiQchHs9hGq4tEyKmVPKSitNmqN,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1035446147",
       "n1NQwT2R8bDpp9URhtyEXSrcdm7b9EBwevd,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1HLya48tasDaMnBpLkd9CwrH6kEuWYnbiw,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1bTSMkk9Axw4JJPdki9y4H5qEa8nZJ8xsP,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1HSLv98ZfZhpDrv7xMP2Bjd6PfAND79AHJ,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1JctpYGdkSZaZgnbz9sCaQC9YgWFMHbazS,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1JD5iug3EgAFSDgGSugmMX7yFkp3K259EV,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1097959568",
       "n1U2RfBPmZTr9eDNWGbYZhq5ahK7nnkozs5,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1az7ZPRnheXVPAZx1UrBaviv9i1fwoo8DA,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1108552833",
       "n1MddT1dPpcgR84Hm6frSRkgr5M1G9ttXbM,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1075419609",
       "n1VGmNA3TxHnRRsvXdTekhR5LeQDRM9ByJi,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1TKr3ko1VfAPnPwvHbRPrRYJ7SFWmJRQis,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1194965551",
       "n1LeSAZqsCG4EtuuDR6T8wbYz3JpKsp3xo3,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1167508769",
       "n1EaZqWVrwxDHqDhXrqX4nethcNZ2sTMCtu,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1053044065",
       "n1Zf8SCheBEw3Xfbv4KQDHS8Tx8ZXBm1zCC,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1111200318",
       "n1S7MtnVjp3JiCRVGQiyopucFHwjXvMRDvk,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1029636848",
       "n1WyhnvnBVFyFG3Gbv7tFPi9vXzTWADoo1h,"
       "n1w1UZ2gWhPe6qp8bLR4wsf5275JCj1dNC1,"
       "1080619077",
       "n1ckTSc1qtmtTfyhjRQTyDRaiAiZAJUtgir,"
       "n22f8VKJ3ymzpAH4iZWJLUEXMBCTv6cbSF9,"
       "1120747042",
       "n1FA49mSfWFkjYHiwhZsg3iVd4eZqPCcTQ2,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1018340167",
       "n1MgC4YefeDaD1wgcmcjg3aMw8v1A66Ngte,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1SENMkgEnnT4yB8H8Pu7TJXcFjfERdhK9B,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1YwUrXvrUx8nNoYn3dN7yU2D69M4kq5oRu,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1M8ZDFxjfaJ1A7beD8miMKZsuXwXe8JYeo,"
       "n1hiWG7Ce8HhTaJGzSJoAaJ9w1CJd7Do2rm,"
       "0",
       "n1SnyYCNcFfcQzMR8atKRvLmFFKi5Yr7f9x,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1115771710",
       "n1PY341gNFWiWGFEoVWGvxzrm31w9YcVmEi,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1ZNSNDMtMuqbyqB3958yuvQwMpuJj3LY83,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1029561548",
       "n1FmyLvES1zu2SadNgAedqULRGKQLgFuF15,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1FCfCQy6d9k9yzKx8qMvbSt9xNMs6YcsmG,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1021579759",
       "n1HVtTZW7V7U4BSXfcWFW33wceGc7R24CqF,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1073180097",
       "n1djtTiVMfj1jA14tfymaC1aKRbT6ok6Gej,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1SBi1hCp37DLa6duGjd2HG3KwAHPCGC8r4,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "0",
       "n1RuykDSbypfEnasQXxmfYiQEafo9xxJQtP,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1150527123",
       "n1T9fbfNbgNe1C8s97uNKgBxD91LUQ4vtti,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1122581474",
       "n1K2MYYg7JfLgUAbWVMp3bHHwGReU4CAHDe,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "0",
       "n1UwuXBVFxV1zJwZvDUia5uL3PsjxNht82m,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1Sayf8Yqif9n69wi9kZDXd1PYQZfnE4wEo,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "0",
       "n1YxyaBzoRogNh72eri7HBGijCAtcHpf9nm,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1158007225",
       "n1Kwt7jqUHcQnpvnsppQfKnNEkxGcFUTwae,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1WMF78Z71ftKUYUNqPhrAKqRYJoqH12VqR,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1Sp26toNq1V9Nh29EHKxkfSRvNEtL7Z91a,"
       "n1w1UZ2gWhPe6qp8bLR4wsf5275JCj1dNC1,"
       "0",
       "n1QGTJJCdvxbyferkrASzhfbmHBWQXqCFx8,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1033013674",
       "n1HTz15qyTHnBCUgbxq8WpM9Bw44EVVJm9a,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1ZsQDS2ofg7Knef87xo8nyNdYPP665v5TH,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1029097185",
       "n1Xz96gDHrrs4ymBgpirav6LbFxZ1p5aRL1,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1130676819",
       "n1TTKMVw72SNTM5o1ZfGihezst4BJ9wQEij,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1WFsXQugai7KDjNoK9k2Lk3jDwTj8ga8yw,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1035477492",
       "n1ZD9dgbsB5kqsgrz5RXyuQH7iGY2GE8AkE,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1176840223",
       "n1Ldtku1LkzmjDQrws9p8r3qEHYpEjQ7Q2f,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1XuvyJeqKjTCc8eFNyuf7PQtqHUD9XB8UG,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1WDiQ27AGGz6f2SxsPaNQ5JTPaXdKoeK3M,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1YEbBHAyHfumQtKYJjfh4BPLhKS472zskb,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1UrTR1PVMyKyCjWje5QxgrDhGqiNGjU77J,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1082322561",
       "n1ViP2uAc3pKPoAZBG9Eq5haFcHzaELpfsi,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1KjEaExsy1MQVRrDkp7uFniP6qAHHm5VLL,"
       "n1w1UZ2gWhPe6qp8bLR4wsf5275JCj1dNC1,"
       "847521382",
       "n1U1U1xZa95gJbByUXKNQ3A6zswYtDxz2z6,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1119398587",
       "n1aveZyHwQuU8Td1JAYbetvsUwQmkSyVjZG,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1SZEyfdpdXjzeQUbi5G5erL9nFiALWpfcP,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1F5GnbTYAgrKfgd3p7JkGMydpFJrQqVnEo,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1144824970",
       "n1Jso8nDBoRCDFg1PRZQHsBjfWJdWZpkyJF,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1113804070",
       "n1MWYFB9Ge2Eh4KbsSc9CLZXef3ACcwzcP7,"
       "n22f8VKJ3ymzpAH4iZWJLUEXMBCTv6cbSF9,"
       "0",
       "n1WTqvVbq9gVgG7NzmfvNg7Dxpa9gqJ4vGR,"
       "n1y395hWWXVUwz7dWG7bjjTrurkfDgNbhhn,"
       "1171447376",
       "n1GV9UScgncwU6KKL9T18mCo2S6uAE69SWs,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "1126029837",
       "n1axFckepP7RSgJ9JDve5EmaYGUs7L6dLy2,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1192235888",
       "n1NZwgrEqoFXbTa2FAvtV6aZC5jNZrNEKsJ,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1YPMjEDMrZhroKmB1xDBhadygwWHC4zTwm,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1141723453",
       "n1XFyqsvCBu4pGP8tFSVvUipmw67tqGAkmm,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1141012942",
       "n1Ltqfvirt7yH2Qzhy1wr2iUZPKhWm2Yhji,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1015396461",
       "n1VjSRU8uwm7jvjZ5u6dCZHoQSG49ERP2Ji,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1020717309",
       "n1L4A1RuKgmdwVQ93NAcYAt2dmjbo8oL63N,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "994388252",
       "n1LARLovXdZ114keF4hRPovFQdaRrTdhjVW,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "0",
       "n1PB6CrnwuLyeYsZHFNoT5x9c8322wBnBf6,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1048339054",
       "n1MCHsH1AwiAtxVdXd33vGK71Qd6MwqMjqq,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1L89WKut3WVfgn2s1keCTFmrjJ5sor4V7j,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1XVMuPmqPBMd2ctz4PjKrmDRjoMXeWxVcb,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1171271047",
       "n1UCDTV6aTDLU9haSY61zofzUyGDrthkPMv,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1074735755",
       "n1T9WMx2FbXSJJRhvFetxCjYvFEpaULyem7,"
       "n22f8VKJ3ymzpAH4iZWJLUEXMBCTv6cbSF9,"
       "1004443274",
       "n1PSDw924XWy7exx8fp5XbkZawCxVQNb2BY,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1105519877",
       "n1XRYjMN5SHu6KepCJa4DZPZX2rDQx6GJY5,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1074047389",
       "n1Sr3QsFdzejg4qngumJWUxP8QaXv3a5V2G,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1XXoWQaucXCNew4cYUhzXzPKAXAQzRPbvt,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1130634203",
       "n1RMFy1gah63To7U9L3hBESeNErmu4sG8Vc,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0",
       "n1acHxCemXNXwKnjWbRzchixrDV63DvxiTu,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1183309827",
       "n1Ugq21nif8BQ8uw81SwXHK6DHqeTEmPRhj,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "1232629956",
       "n1FNWsyLWE88JEAFAg3VtDMiGkU8Q1kf73t,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1195044861",
       "n1KtbBZoh6evjhwELpBJ9vy2u8sADFvfYVX,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1014038821",
       "n1YWZnxoPXAnGZmce2CxVscZNrEtGUxmZ7B,"
       "n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39,"
       "0",
       "n1cgGKaCF9jFW7KXUnyPV3vN5dTwgz7rhUc,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1147162587",
       "n1X2SXyEKej7GZgAreXDCkiT59qaYKBDcYi,"
       "n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB,"
       "997704071",
       "n1bgExuWHyHWibA8BtbYfq9T9gVwu4vtEjC,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "1184148227",
       "n1X6nw1SUU9d6QT1CpyUt7yn94F2jtpTv68,"
       "n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H,"
       "0"});

  for (auto &line : acc_contract_votes_str) {
    auto ret = split_by_comma(line, ',');

    auto source = ret[0];
    auto target = ret[1];
    auto source_bytes = neb::bytes::from_base58(source);
    auto target_bytes = neb::bytes::from_base58(target);

    auto mem_int = std::stoi(ret[2]);
    auto f_x = *reinterpret_cast<neb::floatxx_t *>(&mem_int);

    auto it = acc_contract_votes.find(source_bytes);
    if (it == acc_contract_votes.end()) {
      std::map<neb::address_t, neb::floatxx_t> contract_votes;
      contract_votes.insert(std::make_pair(target_bytes, f_x));
      acc_contract_votes.insert(std::make_pair(source_bytes, contract_votes));
    } else {
      auto &tmp = it->second;
      tmp.insert(std::make_pair(target_bytes, f_x));
    }
  }

  std::unordered_map<std::string, std::string> actual_ret(
      {{"n1w1UZ2gWhPe6qp8bLR4wsf5275JCj1dNC1", "8c36f43f"},
       {"n1y395hWWXVUwz7dWG7bjjTrurkfDgNbhhn", "344aa442"},
       {"n1n5Fctkjx2pA7iLX8rgRyCa7VKinGFNe9H", "9cf70f45"},
       {"n22f8VKJ3ymzpAH4iZWJLUEXMBCTv6cbSF9", "92806442"},
       {"n1fPbCgC2Gt4ABcxdeBVbYPv6JHNVJJ2G6q", "bf984543"},
       {"n1yCUCFCpLzjY2E7iBEBNtQYcmuLBzQoMxB", "124b2d44"},
       {"n1hiWG7Ce8HhTaJGzSJoAaJ9w1CJd7Do2rm", "25d0d31e"},
       {"n1zUNqeBPvsyrw5zxp9mKcDdLTjuaEL7s39", "7e229d44"}});

  auto expect_ret = neb::rt::dip::dip_reward_v2::dapp_votes(acc_contract_votes);
  for (auto &it : *expect_ret) {
    auto ret = actual_ret.find(it.first.to_base58());
    EXPECT_TRUE(ret != actual_ret.end());
    EXPECT_EQ(ret->second, mem_bytes(it.second));
  }

}
