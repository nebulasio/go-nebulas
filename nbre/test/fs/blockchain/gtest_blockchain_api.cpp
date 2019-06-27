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

#include "common/nebulas_currency.h"
#include "fs/blockchain.h"
#include "fs/blockchain/blockchain_api.h"
#include "fs/rocksdb_session_storage.h"
#include "test/fs/gtest_common.h"
#include <gtest/gtest.h>

TEST(test_blockchain_api, get_block_transactions_api) {
  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());

  auto txs = bab_ptr->get_block_transactions_api(12802);
  // clang-format off
  std::string transactions_str[] = {
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1TqDPDNyWgz8EmjhWH33GXihbmuhaEL3HT,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1axo85s4uV1YCFfoSasr1DtNKY6SPkUcix,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1TVSb81AoheQ9HamGoG97S5ALQB4d7Hz5o,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1SypkofeeepccjqSMcYTdh6ekvYk549uLa,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1GYbe5epy4Vv4wFoun9w8MuN7rCVD8bLzy,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1bAqofrt6nEgQNoheuG3Xs6YcAws6nuEvU,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1ZcWweVmJ7UdP3VocfcmsKUHQGq73BQt6A,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1K5oPa3DzcABBWDcARh59kCN4RuTMGVNEf,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1cKsA4KbJ2vxhUAd7Si2fD9wXd9XuCj1XC,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1ZZQpYrFap2gJy3ikymsf4MT86RaxgJuJ7,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1LZPfTptU1d3rLqw7SKChhifHngHetdsae,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1dJZyEdtXmT6d9Ttn7ypr9gRYnnQqQ3BXF,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1FrXEYbjVDhKTNewDmfXbXV9mKdCBoGGsU,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1TsAszVEVX8D68zMtuB21MfnjrfUaWLwe9,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1aJDHoXxrvozWVhjWRJ2jEVhCAwK5pDKfX,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1R3H2CrV1LWvoincqvGpuj5ccK4TG55QEP,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1NtJFCVTmusYQ5rE59aQB5Ds5CPcZf2yuo,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1VMjtBrs67oKaxSTv3Z1hxhZodmyRAQe2g,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1PcWTCjxBWBKzqjsPyjBPnppQseBCx51bC,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1F198tRHFMj9fVDbdoYhPQrm6Nabg6qTzr,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1Ry5pmuhGeXWxGuZFCKYmuUFMBBwGhBLD9,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1FEhiEEnSLosnU9wxYWgUGPc5DpCuJRwsW,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1KFa1mznizyeuQkZVajfufJKW3CmuipsdN,binary,1000000000000000,1531710285,20000,1000000",
    "12802,1,n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m,n1ULxuL5W7EkkyRTmCKtaAebc24AuFFriav,binary,1000000000000000,1531710285,20000,1000000"
  };
  // clang-format on

  for (size_t i = 0; i < txs.size(); i++) {
    auto &tx = txs[i];
    std::stringstream ss;
    ss << tx.m_height << ',' << tx.m_status << ',' << tx.m_from.to_base58()
       << ',' << tx.m_to.to_base58() << ',' << tx.m_tx_type << ','
       << tx.m_tx_value << ',' << tx.m_timestamp << ',' << tx.m_gas_used << ','
       << tx.m_gas_price;
    EXPECT_EQ(ss.str(), transactions_str[i]);
  }

  txs = bab_ptr->get_block_transactions_api(0);
  EXPECT_EQ(txs.size(), 0);

  txs = bab_ptr->get_block_transactions_api(1);
  EXPECT_EQ(txs.size(), 0);

  txs = bab_ptr->get_block_transactions_api(10);
  EXPECT_EQ(txs.size(), 0);

  txs = bab_ptr->get_block_transactions_api(12785);
  EXPECT_EQ(txs.size(), 2);

  txs = bab_ptr->get_block_transactions_api(13697);
  EXPECT_EQ(txs.size(), 7);
}

TEST(test_blockchain_api, get_account_api) {
  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());

  auto block = bc_ptr->load_LIB_block();
  auto LIB_height = block->height();

  // std::unordered_set<neb::address_t> s;
  // for (neb::block_height_t h = 1; h <= LIB_height; h++) {
  // auto txs = bab_ptr->get_block_transactions_api(h);
  // for (auto &tx : txs) {
  // s.insert(tx.m_from);
  // s.insert(tx.m_to);
  //}
  //}

  // for (auto &addr : s) {
  // auto ret = bab_ptr->get_account_api(addr, LIB_height);
  // auto balance_wei =
  // neb::storage_to_wei(neb::string_to_byte(ret->balance())); std::cout <<
  // "item_t{\"" << addr.to_base58() << "\"," << LIB_height << ",\""
  //<< balance_wei << "\"}," << std::endl;
  //}

  typedef std::tuple<std::string, neb::block_height_t, std::string> item_t;
  std::vector<item_t> v(
      {item_t{"n1sNXeyJhsnknTfYiRQtFsS2epwUp3Ls9KH", 23078, "0"},
       item_t{"n1xaW17cS9DTzPksmPHzhaoaC9yGfCFQy4C", 23078, "0"},
       item_t{"n1pohg1UWjFFEKb9XnqGRdscabcBrT7WbaF", 23078, "0"},
       item_t{"n1Fw763xUpPqkiWU1zsqUVReG1ETzhJp9eU", 23078, "989979893000000"},
       item_t{"n1TScMSkMacEFUX2d37Cvxbs3hMSQzqLow9", 23078, "989979893000000"},
       item_t{"n1ZKNtQZKCgdJfWmvzeR1vDuU5JErZXivYC", 23078, "989979893000000"},
       item_t{"n1TWLPisUy6UUsxTKPevyQEgq3Jew9rv9ww", 23078, "989979893000000"},
       item_t{"n1PZ3zgP8yHxGt5qsGR7N4G6SLoD2rBF1C2", 23078, "989979893000000"},
       item_t{"n1Emg4q43xrjySKtSU8AkW1y6WGT2KcfC9x", 23078, "989979893000000"},
       item_t{"n1KrzDxim8jjtBdGGJU8F3HMk79fVmQWtmi", 23078, "989979893000000"},
       item_t{"n1N25HeoB1Uci1VVLWUyXYEqTSVAL5vFjEV", 23078, "989979893000000"},
       item_t{"n1LTLN7VGToigJLRxfvti3v1ssBiDd3Uee8", 23078, "989979893000000"},
       item_t{"n1bAqofrt6nEgQNoheuG3Xs6YcAws6nuEvU", 23078, "1000000000000000"},
       item_t{"n1WkTyCqxWDPoEM96eoNakKGjc4wazGwB3A", 23078, "999979878999999"},
       item_t{"n1Uc7eZzPtt9mVatvEvNhdqVHPoF7NFDANJ", 23078, "1000000000000000"},
       item_t{"n1nT22NaMhH9GfQCauKdXQ9CruQ72uuosmR", 23078, "0"},
       item_t{"n1JKY3kFRuStzoUxMr2qpPfFJ3U3GkY3kNd", 23078, "999979878999999"},
       item_t{"n1KWEapwVkVx5cqPrV8f9WtPdHvZTGQYK5G", 23078, "989979893000000"},
       item_t{"n1L3Ts1zcB57pLuQguFM3W2ejeb7g2rUgVj", 23078, "1000000000000000"},
       item_t{"n1G93Qgpr7eXEj5CGb2ZhTyfCJggYi2wcfR", 23078, "1000000000000000"},
       item_t{"n1cosmXCR1VycUu8HV9V6sHbTfcurEAzYCZ", 23078, "989979893000000"},
       item_t{"n1N1o1N7uwFEQ8xhdBRQ4nNwLJBwTDtoHvG", 23078, "989979893000000"},
       item_t{"n1XJFYd9s2CMsDso1VwS3a7VfAgcUu5xcCy", 23078, "1000000000000000"},
       item_t{"n1FNtWBQmsxQ4o9eFncm4grwmDoY7miUEyC", 23078, "989979893000000"},
       item_t{"n1WRtHJrGmrKyphArNtUQ3sTC3P5GU9pqoh", 23078, "1000000000000000"},
       item_t{"n1SeopPqcymKGJGtSTQydJ17Y7VzM4AtBto", 23078, "989979893000000"},
       item_t{"n1EfEYK5e3cCHnSxeV87WCcqCwmiDhgXM2C", 23078, "989979893000000"},
       item_t{"n1YheQGFaxTnDRiRLi2v7axqjm2fnwRT2tB", 23078, "1000000000000000"},
       item_t{"n1djc5fgkkQC4PETodp69aDBqmyT9kVskgH", 23078, "1000000000000000"},
       item_t{"n1K5oPa3DzcABBWDcARh59kCN4RuTMGVNEf", 23078, "1000000000000000"},
       item_t{"n1XEY4fLnSvHsr6nitPecRPDagZSxyt6jjc", 23078, "989979893000000"},
       item_t{"n1ZZQpYrFap2gJy3ikymsf4MT86RaxgJuJ7", 23078, "1000000000000000"},
       item_t{"n1YxeasTe59mGum3HeVKPZSUHWnFhFXRTSr", 23078, "1000000000000000"},
       item_t{"n1d9oyba2Vxh353pH4q4iy4zak4s4bS8gSu", 23078, "1000000000000000"},
       item_t{"n1FKncsEWdgyxA5rpoh5oKHpH43WECUE7L8", 23078, "989979893000000"},
       item_t{"n1QvRhgDtx1bFAFEN39MaQZp89o4sh1BAG9", 23078, "989979893000000"},
       item_t{"n1Rv51xiqjXpneANYsemVaPwvyCPjcAywMU", 23078, "1000000000000000"},
       item_t{"n1M9SMLX72hnCyLpzNyert2p6uWa4YXqQVq", 23078, "1000000000000000"},
       item_t{"n1dSs99ei9HEZ2uNwP4i4yXLmDrPxAehWF5", 23078, "1000000000000000"},
       item_t{"n1G4rf3uqGYfZhS1QvLYAu6tpQhJTWH4b9n", 23078, "1000000000000000"},
       item_t{"n1ZcWweVmJ7UdP3VocfcmsKUHQGq73BQt6A", 23078, "1000000000000000"},
       item_t{"n1ZUWS4nPDbbZXP2HCeAPj5tys6KEVBcADe", 23078, "1000000000000000"},
       item_t{"n1T94a4ko9FBv3LCxGnJLbxBabDyiX2FXFu", 23078, "989979893000000"},
       item_t{"n1ScXRMuVkbuYZmzv1devivJkTbxHmnXCen", 23078, "1000000000000000"},
       item_t{"n1TVSb81AoheQ9HamGoG97S5ALQB4d7Hz5o", 23078, "1000000000000000"},
       item_t{"n1GnosAeb1ifiWdJEfnFq17ybHSrq981Vkk", 23078, "1000000000000000"},
       item_t{"n1P32i2vivgS8yWVTf5bCToEz7CgrAvrdWx", 23078, "1000000000000000"},
       item_t{"n1dkuk9d8X9JiDT5rS6KS7CTiGeZJp3oi8g", 23078, "999979878999999"},
       item_t{"n1K2GPoZSA4k9m1vfix8w238igNf8MMCXCc", 23078, "1000000000000000"},
       item_t{"n1bp6LSrhaEhSXw9mEXbKsormF4Ff2yfnkS", 23078, "989979893000000"},
       item_t{"n1G4VHZP66Q6rptxr2Kf5oHSCX1BRhp24XP", 23078, "1000000000000000"},
       item_t{"n1XEScWcFmrWgbidJxbHGtWwTFTf7g29HQn", 23078, "1000000000000000"},
       item_t{"n1NQRCT51iRNdHrtUCf2LdwncHmhFSXcHMf", 23078, "989979893000000"},
       item_t{"n1Q4uHGmFvxeqt3iCzZnzmSHnnKkW6inK2d", 23078, "989979893000000"},
       item_t{"n1bSaYVEyitJpbuZTQwyoqRyud85gBBxmc5", 23078, "1000000000000000"},
       item_t{"n1koYepMB2DD6eFu6xfZvXQ7NstQ6FXeb69", 23078, "0"},
       item_t{"n1PT9zhGnpy7f681pEnFfuqvRjtW1ZMQYSt", 23078, "1000000000000000"},
       item_t{"n1LUuBuHAhRmE76En2em5Ynx3ToKZ5dyVcn", 23078, "1000000000000000"},
       item_t{"n1H8PZBuuW9vqQE66xm7hFmV8Go8EYN5g16", 23078, "989979893000000"},
       item_t{"n1axo85s4uV1YCFfoSasr1DtNKY6SPkUcix", 23078, "1000000000000000"},
       item_t{"n1HJ61dccMLZyWisy3v9b9zvLi7xzq1YN42", 23078, "989979893000000"},
       item_t{"n1btSQSQF9shq5WNjhsncJYHap2UoXEVYbL", 23078, "1000000000000000"},
       item_t{"n1S2RGeS9ExfYRAdyu2XgFgoP8gKXeXTwrp", 23078, "1000000000000000"},
       item_t{"n1ZjRzg5GiNif4yKA2nJz8v4iXPJb7SZY2a", 23078, "1000000000000000"},
       item_t{"n1HdywvkBPofiatoT7T29VFBYNmqeFVsc8g", 23078, "1000000000000000"},
       item_t{"n1bYxadMFrj2b71cccHyq38SyULifB7NBty", 23078, "989979893000000"},
       item_t{"n1Wphh3UA3GTwBq52EYdJAExNcLFnqeMyuE", 23078, "1000000000000000"},
       item_t{"n1YPGzQvq1fcxrR4S8gZxuY8NxPBAxZ2ryV", 23078, "1000000000000000"},
       item_t{"n1P2SGVB9jCj4aiCB3SY2dth82PKvURgbuJ", 23078, "1000000000000000"},
       item_t{"n1XdJr6ce9QDL2coaxVYK2cdEYHVuYcCNXZ", 23078, "1000000000000000"},
       item_t{"n1GCFQ8ooBThGWuzxWyrmxHKi6LDb36BBgJ", 23078, "1000000000000000"},
       item_t{"n1QGHCxxQP7VKHHPusoL8gndY2aEvfzqgXG", 23078, "1000000000000000"},
       item_t{"n1TqDPDNyWgz8EmjhWH33GXihbmuhaEL3HT", 23078, "1000000000000000"},
       item_t{"n1Jsgn6mSnyQmRtAHkLfmJQ6WwvN7znvZ3m", 23078,
              "9999999789994368930000000"},
       item_t{"n1QfUeTQ7UXjCmz7Jj3wGo2U7YVMq368wFh", 23078, "1000000000000000"},
       item_t{"n1Z1E7zdmXp3Pkns8JqfDco18ApM63AMM4Y", 23078, "1000000000000000"},
       item_t{"n1ZscH45VkLuCKqLAwVyMBSZau1AK3d1So8", 23078, "989979893000000"},
       item_t{"n1K9Y8ofh8R6HpRjX8mkoRj42GGeBDt2Gck", 23078, "989979893000000"},
       item_t{"n1ZukrvE6xxfTQEbSoiD295reNMJmuSaJiC", 23078, "1000000000000000"},
       item_t{"n1R9YdiHz8mpfKod1gXGrhyC2YutratH4qK", 23078, "999979878999999"},
       item_t{"n1X9DuRHHYM5ShsupRPKdJycURgquQLiZ8K", 23078, "1000000000000000"},
       item_t{"n1VA4KRgVTUNQBgNHnC2ryARNgFMY6SbRB8", 23078, "1000000000000000"},
       item_t{"n1UQm7Gxpb6USxMxs3VuimML7XEke6xQ4ff", 23078, "989979893000000"},
       item_t{"n1FmrbppsF5xyahWMUw7jJtqCPbPUdff9sA", 23078, "999979878999999"},
       item_t{"n1WWCw93edj9ZNwjStLJ4DJz4ZdoDxKgdm9", 23078, "1000000000000000"},
       item_t{"n1Kk5KD5Do7Zo5oUrLe9yw7EeXk2nDrYo2W", 23078, "1000000000000000"},
       item_t{"n1MGjXGSEfmBGr7bV9vPWEajvfojRGurwtb", 23078, "1000000000000000"},
       item_t{"n1cX3TKhHCUyPyi4o4ByqJ77hL5XYFkKRzY", 23078, "1000000000000000"},
       item_t{"n1NXKcSXsJouLcj9ycyLdkJiT3EP6dkGFST", 23078, "989979893000000"},
       item_t{"n1TMeLpTDWoNgCgGXTp8esb7eCfbzqkTyYK", 23078, "1000000000000000"},
       item_t{"n1UzMM6pwVfYTVpyemG2bYCx32eFRkWYMYs", 23078, "1000000000000000"},
       item_t{"n1R4MVhHfwk9gnxrE5aYhSrkMhkSjnDL3ZS", 23078, "1000000000000000"},
       item_t{"n1GvE9Gu2W6ypiiUqoXpaPDBkFy5D2H32uy", 23078, "1000000000000000"},
       item_t{"n1GYbe5epy4Vv4wFoun9w8MuN7rCVD8bLzy", 23078, "1000000000000000"},
       item_t{"n1SnWpoukLndmq2zp2ftQgsZJgJ1cJkkkaV", 23078, "1000000000000000"},
       item_t{"n1RxWmhBR2Rz8iKTq5wb3dy9DCMzjN59gMw", 23078, "1000000000000000"},
       item_t{"n1EdcEeJakeqpTLMRHGbt3Mkbz6Kbwy6Ymw", 23078, "989979893000000"},
       item_t{"n1KR9wa5fUvbSzp17EC522dhuc7ySkAgaP9", 23078, "1000000000000000"},
       item_t{"n1VJnC9BmHx7WmGmHav1ZLaRERoJa8cjqcD", 23078, "1000000000000000"},
       item_t{"n1R3H2CrV1LWvoincqvGpuj5ccK4TG55QEP", 23078, "1000000000000000"},
       item_t{"n1Sf5HASQ9JLwe72wM8nYWTQeLS7YcNkv4h", 23078, "1000000000000000"},
       item_t{"n1TgMPpj92Nj8c6pCzu3DsGebBx9ke4tHx3", 23078, "1000000000000000"},
       item_t{"n1SypkofeeepccjqSMcYTdh6ekvYk549uLa", 23078, "1000000000000000"},
       item_t{"n1YTZ41uJrb6HRPjz7MeK1i6J6GhSrnNA2o", 23078, "989979893000000"},
       item_t{"n1PpiK4XKvYxx2AovYmULDfB23toJZAcFAw", 23078, "1000000000000000"},
       item_t{"n1JEx5SkwuosRvwWbK4vpf7t9vXyEgdSdur", 23078, "1000000000000000"},
       item_t{"n1ZQ5VDioNTj98JZ4nGyNBnpWSyvGRRmgdK", 23078, "989979893000000"},
       item_t{"n1Mxf1fpXAq4aM9juZy6w5JuqxZRyEwyQBD", 23078, "999979878999999"},
       item_t{"n1NvgWLL8S1bgyMYE6c6sskFaFzwwqf42QE", 23078, "1000000000000000"},
       item_t{"n1cX8kkoKPUJNS4d6CSmzFRoWqVuJd3YyrY", 23078, "1000000000000000"},
       item_t{"n1KDjSqEwehu1mzg2ZT7PrMbD57VCyKZonr", 23078, "989979893000000"},
       item_t{"n1EyWJDYA38i7Xqc7mtvpSQpyT3kLXcr57j", 23078, "989979893000000"},
       item_t{"n1YWWM3z1p3CTh4yJ8d95iriP9PJkh44hZZ", 23078, "999979878999999"},
       item_t{"n1P3P3F4MQZsFRteaCGwY3bryCTjnmKzU4c", 23078, "999979878999999"},
       item_t{"n1cHFuGji4HEhZJ6R6cNaV44vtUSthbWyu6", 23078, "1000000000000000"},
       item_t{"n1FM45GD5Xq5L4CJjJhKf1SG5551asHrBWC", 23078, "989979893000000"},
       item_t{"n1XfKEuXjRKxEnhxh8Gy7qXHhgUuDPWcBZ1", 23078, "1000000000000000"},
       item_t{"n1KqEJCN4cWE3bagyLcHswAdtNMneK5toV4", 23078, "1000000000000000"},
       item_t{"n1dnhFjQD5rKjnSuaVSQTYbmB6KzzaYEVsT", 23078, "1000000000000000"},
       item_t{"n1XxNHvhf2DJfpWGi6XKqi5JMfh1iu1zDqb", 23078, "1000000000000000"},
       item_t{"n1M4VzqiMHc8GZU5NqWU36YPiCuG9HUtDEy", 23078, "989979893000000"},
       item_t{"n1acBGAeHGwj42CM5t7dqjb7CQWrZELordj", 23078, "1000000000000000"},
       item_t{"n1P4YQmHP2FLGNrzoehuFYp9gbPaRLNbNMY", 23078, "989979893000000"},
       item_t{"n1T2ksvUUqw2WMHfdpvwBajsXQ5sEE7pape", 23078, "1000000000000000"},
       item_t{"n1e22bqyXSFpbKXUysKRemVaXyBDqEPrA4S", 23078, "10"},
       item_t{"n1Y3RsrHkrRJGdrFnvSFwd9rkUUJrJZe7TJ", 23078, "989979893000000"},
       item_t{"n1NY3KS8rGX382VURgabYu9AzinP85wnvBL", 23078, "1000000000000000"},
       item_t{"n1H7F5VK4KnysZEP1vH9iyFVR4nFfdKww7v", 23078, "1000000000000000"},
       item_t{"n1RDZmj4pzzTaj2geAApoZZovYgoFcSt69L", 23078, "1000000000000000"},
       item_t{"n1cKsA4KbJ2vxhUAd7Si2fD9wXd9XuCj1XC", 23078, "1000000000000000"},
       item_t{"n1LZPfTptU1d3rLqw7SKChhifHngHetdsae", 23078, "1000000000000000"},
       item_t{"n1bPQYJdpVVdzvfj9dP899WaPRjDquYrRdy", 23078, "989979893000000"},
       item_t{"n1dJZyEdtXmT6d9Ttn7ypr9gRYnnQqQ3BXF", 23078, "1000000000000000"},
       item_t{"n1TsAszVEVX8D68zMtuB21MfnjrfUaWLwe9", 23078, "1000000000000000"},
       item_t{"n1SzPNtdTuKhAfqyo1Cdy9VhYWYz4Z7fW8N", 23078, "989979893000000"},
       item_t{"n1NtJFCVTmusYQ5rE59aQB5Ds5CPcZf2yuo", 23078, "1000000000000000"},
       item_t{"n1aJB5pEZ9i445i6eyxAV6UkHjoeZpUvLCW", 23078, "1000000000000000"},
       item_t{"n1VMjtBrs67oKaxSTv3Z1hxhZodmyRAQe2g", 23078, "1000000000000000"},
       item_t{"n1aWaeMwtxVTsrxWwSuogvj3FugJNp99vhZ", 23078, "999979878999999"},
       item_t{"n1FrXEYbjVDhKTNewDmfXbXV9mKdCBoGGsU", 23078, "1000000000000000"},
       item_t{"n1KxTCyE53Px3kuVb6LFyzxtchNLCdAtCUV", 23078, "1000000000000000"},
       item_t{"n1PcWTCjxBWBKzqjsPyjBPnppQseBCx51bC", 23078, "1000000000000000"},
       item_t{"n1F198tRHFMj9fVDbdoYhPQrm6Nabg6qTzr", 23078, "1000000000000000"},
       item_t{"n1LzjZ7Q6xto4wziJdiUu5kyrwQ1BWtW5Kp", 23078, "1000000000000000"},
       item_t{"n1GZ74L9J7TKf6tDFBheJtBYCzBrzmvJWnt", 23078, "989979893000000"},
       item_t{"n1FEhiEEnSLosnU9wxYWgUGPc5DpCuJRwsW", 23078, "1000000000000000"},
       item_t{"n1ULxuL5W7EkkyRTmCKtaAebc24AuFFriav", 23078, "1000000000000000"},
       item_t{"n1TPJzzi5N4Nmrhoj3YgchJE5ayaTkGEY3E", 23078, "989979893000000"},
       item_t{"n1UjQN378CA1SN11b4Nnacc2KzMP8dc6qpW", 23078, "989979893000000"},
       item_t{"n1NRa6oBgstmrYarxZyRy9eNeB1iVh9bVM7", 23078, "1000000000000000"},
       item_t{"n1FDJhTL1VzzmjSqYT1PKFfd9YuvixSVi8H", 23078, "989979893000000"},
       item_t{"n1FNCwZq12JAPNNaosMjQRUCEeo9E3vMzkY", 23078, "989979893000000"},
       item_t{"n1dDsPV3NEgxL2ARM9uJ3MYhsyc41NFSTMk", 23078, "989979893000000"},
       item_t{"n1Z13rnp5ogvLdGC5tMfhHqJrJykQ9H2mHb", 23078, "989979893000000"},
       item_t{"n1P3wk4M7DtSeeDqznaqaxyGYs3uHv7NX7m", 23078, "1000000000000000"},
       item_t{"n1F9VjU9FgUCRgXCPEX43bdcZGdM5h4E2zP", 23078, "989979893000000"},
       item_t{"n1ces2mwyoBWZ2z1HsNVciaU33Ft4KMUznM", 23078, "989979893000000"},
       item_t{"n1L2wARA12CppMu6RdgeczTSAUsneYwe1QQ", 23078, "989979893000000"},
       item_t{"n1T7fvSeUuU8xbmuZzbAyFY7hxb7WnFrkfs", 23078, "989979893000000"},
       item_t{"n1aJDHoXxrvozWVhjWRJ2jEVhCAwK5pDKfX", 23078, "1000000000000000"},
       item_t{"n1YRGnHTp3k7bFGfu8tbB2ucfbaui54ZymL", 23078, "989979893000000"},
       item_t{"n1JCZqmrbq3Te9u8WoX5WbbsaePWKiehZjA", 23078, "989979893000000"},
       item_t{"n1VqbtMhz5GeABcUj18XHM5z7bmkBzQVfV6", 23078, "989979893000000"},
       item_t{"n1NAUVx3LZGhgjVnPUMzdhMfknouBdMdTJP", 23078, "989979893000000"},
       item_t{"n1dZqX5jp5qK62acGJWH5Nzb9gFKoiYTkmF", 23078, "989979893000000"},
       item_t{"n1EnQEzxkqge8XsH7uFaVY5ahsM3srDhPhP", 23078, "989979893000000"},
       item_t{"n1SzWZRdJcPWu4WiVScSqq7U5GHRwaRB72i", 23078, "989979893000000"},
       item_t{"n1r6XsVSD7eKrNWqpBsqXAVq3ZhVu8pWQ2m", 23078, "0"},
       item_t{"n1UmJLpRyVwTnogHeRfcD4NkYS29PDarYPh", 23078, "989979893000000"},
       item_t{"n1QPsMhERyMEFDY2q2btfgAvf4u1gdUoKDV", 23078, "1000000000000000"},
       item_t{"n1bt9DHx2NvGNdAxRLX2qiviVrUtPMVZVSh", 23078, "989979893000000"},
       item_t{"n1MjUZzvsb8KJM1vdrM5KKVCaP8FvRqNoba", 23078, "989979893000000"},
       item_t{"n1Pxp1VwVwu2VR2GHbhkotSaPSK7mG7YJt9", 23078, "989979893000000"},
       item_t{"n1xfspHUmwzvx5xZJQpQ2xC52NzRhxekkFr", 23078, "0"},
       item_t{"n1Hzq6X2XJ3fysspkpf95FBdjpnGvFrk7bY", 23078, "989979893000000"},
       item_t{"n1Khb8a3qTVtKfaVbruF3m3r1MNtQu9GcaY", 23078, "989979893000000"},
       item_t{"n1QS5bLHeLA2UX8JvJwMECFPt2jiPYpX5au", 23078, "989979893000000"},
       item_t{"n1L8pSmvw3fCVLKWCUHX5SxrCAQkUzn4q47", 23078, "989979893000000"},
       item_t{"n1aRyqFrChmS546xuivmf5tpN4FDS3J78dG", 23078, "989979893000000"},
       item_t{"n1Ry5pmuhGeXWxGuZFCKYmuUFMBBwGhBLD9", 23078, "1000000000000000"},
       item_t{"n1WwPXtRnateipBKAzqS6HDZacgpRLW3eKY", 23078, "989979893000000"},
       item_t{"n1VvYG1d8dyi8nN1hNP7Enh49hcC24Mxnb4", 23078, "989979893000000"},
       item_t{"n1WQbuEB7vTznG2Wpo5syPaa8QwGtcp6iZd", 23078, "989979893000000"},
       item_t{"n1XNxPfnSwDcx9VUXAhFhtgPPiweX4c1Njf", 23078, "1000000000000000"},
       item_t{"n1ZSgHeB69iko6bFd8Wd1dGCWwvLksSYHJy", 23078, "989979893000000"},
       item_t{"n1L6EaS62T2KbvMfHbAsWK7pN3UKEABDWyA", 23078, "989979893000000"},
       item_t{"n1aPy22TQo9Ac9NbSbceQv9bLhhHwrnLpY4", 23078, "999979878999999"},
       item_t{"n1GwxMufp9yK26Jttzb2t4nFjmur3o7metA", 23078, "989979893000000"},
       item_t{"n1ZRrr7KcypiieJCAFgxSy6xP7XG1ug4CkS", 23078, "989979893000000"},
       item_t{"n1YFj55m8uoQk3xqsSVxxyRbUMo6zsjpNDE", 23078, "989979893000000"},
       item_t{"n1cjAvpSiCdbSU1hDnepSoDj99FXj5aiQFE", 23078, "989979893000000"},
       item_t{"n1Kmg4YxqE7HFkJedMFwqpDhmvJf5Et43g5", 23078, "989979893000000"},
       item_t{"n1MqvAvVaAsk4Ky51Bh5oAZ21CFrXx5zt6v", 23078, "989979893000000"},
       item_t{"n1KU1uGSZpWcHcoYaFAcSqJsm4AoYCH5LkK", 23078, "989979893000000"},
       item_t{"n1Xns2PvNjXS4MR9KSLenvwRx59E3C7392T", 23078, "1000000000000000"},
       item_t{"n1EeF8L7oirH95sZpf6fPKuVJseNPmhdhm3", 23078, "989979893000000"},
       item_t{"n1mg8VTQJZ4kGDnDrggQnJzn4v5bDpo8bzy", 23078, "0"},
       item_t{"n1cVPpEcvoxX3qyvXqRBso3UdhM8NJxs8vd", 23078, "989979893000000"},
       item_t{"n1KSo7ekbZpm6gEVvHcgtQ9tJU5st4F9dXK", 23078, "989979893000000"},
       item_t{"n1mWX7JAbzZUuDjKB61NF5aPCADrY3cvkst", 23078, "0"},
       item_t{"n1Gx1hR3YfLP5Fc8kKJXXEs561cjvVaFGWv", 23078, "989979893000000"},
       item_t{"n1L1nt5cK6pWS8C2wFzFjTMVSzF8JRvCxYE", 23078, "1000000000000000"},
       item_t{"n1Q6P32NfkJ8wqrjQR8oyckvXB4mJpJpwvk", 23078, "989979893000000"},
       item_t{"n1dpMTzFXXRkEBE37nzXvhSjQ7KXGqL5Y3N", 23078, "989979893000000"},
       item_t{"n1HnVjj7mDZDx9G3ybMfScgSUscwV6V8KL3", 23078, "989979893000000"},
       item_t{"n1R8x8CSMsBLXs6uxm8KVKywE6S8aD88BkD", 23078, "989979893000000"},
       item_t{"n1cUrg627XmKPccVak3gF1aakHqAdcB1PF5", 23078, "989979893000000"},
       item_t{"n1bqWwAYrdxuViBCeA4LcnMe4xYUWXYtZp5", 23078, "989979893000000"},
       item_t{"n1KFa1mznizyeuQkZVajfufJKW3CmuipsdN", 23078, "1000000000000000"},
       item_t{"n1cYsED3wzLbqeNnde7cPrt9EwVqu4BuSfY", 23078, "989979893000000"},
       item_t{"n1STTN5DndkMWops1MWKgY3Wf1YHyrXEYMW", 23078, "989979893000000"},
       item_t{"n1NRhQuWAZtHqBEjYz3UgyoqvNa7SKWQAdB", 23078, "1000000000000000"},
       item_t{"n1bUH87StyKnQnzBKti6x5LJwRpQ5WEBkgR", 23078, "989979893000000"},
       item_t{"n1EzCQzD9hg4U5rNdExeKMKbY1r6ZRaqhsQ", 23078, "989979893000000"},
       item_t{"n1dYJ5DPv9acCKS6wfEhFJtN6gdXpJoymeY", 23078, "989979893000000"},
       item_t{"n1F15rj8JhTeN2bPLPckh6EHYeLjkyz1aBB", 23078, "989979893000000"},
       item_t{"n1Q7Yk1SPjqeSZKoNqwKg8WPqgj2etDCrnK", 23078, "989979893000000"},
       item_t{"n1MxBZwT8kJLsXtkboMrWGJgSuDWXntsVq3", 23078, "989979893000000"},
       item_t{"n1RCbuNRbVuxNhH2JZhwnBGE3Gzwd6HHXvD", 23078, "989979893000000"},
       item_t{"n1mLSCnaEu2EVykhzGvSug9vf35di213KY2", 23078, "0"},
       item_t{"n1QLMmX8JEncCR2jj4TYN3iso8kJgcRnTs8", 23078, "989979893000000"},
       item_t{"n1bvDYkFcxV8gkgriYqyHqGVa5Z5vpXAfsp", 23078,
              "989979893000000"}});

  for (auto &item : v) {
    auto addr = neb::bytes::from_base58(std::get<0>(item));
    auto height = std::get<1>(item);
    auto ret = bab_ptr->get_account_api(addr, height);
    auto balance_wei = neb::storage_to_wei(neb::string_to_byte(ret->balance()));
    EXPECT_EQ(std::get<2>(item), boost::lexical_cast<std::string>(balance_wei));
  }
}

TEST(test_blockchain_api, get_transaction_api) {
  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());

  auto block = bc_ptr->load_block_with_height(13692);

  // for (auto &tx : block->transactions()) {
  // std::cout << '\"' << neb::string_to_byte(tx.hash()).to_hex() << "\","
  //<< std::endl;
  //}

  // clang-format off
  std::string tx_hash_str[] = {
    "e639f746ba6153b9c51a467b318910b9acb363ff80862d1aba1de32d6bfed129",
    "4876062637dc879a2d5c7c6a6fe348e1005c46e9ca8539960ba50515cf846102",
    "4c1a3493802782c83a0d4bbae2bb010f326bd7b143efa3416d7fe7361b4b81c0",
    "f3e354697714afcdf3db4c069c5d13801d44b22ac36d3b916ea7a2c1fe843a4c",
    "9d45d564dbb378818eb8932ba9507147fb98b51daa1e78e42ae6856590c0eced",
    "33bf4a02769c409c9f809ca0d1540c5e3ed64dc0181def1ba1b10901d9324b2f",
    "eff0e7a9f4572dc4c0562b28e4ed7bf37c63886e7cc43c1b068ecae4d5c007e7"
  };
  std::string transactions_str[] = {
    "n1UjQN378CA1SN11b4Nnacc2KzMP8dc6qpW,n1mg8VTQJZ4kGDnDrggQnJzn4v5bDpo8bzy,1,1000000",
    "n1F9VjU9FgUCRgXCPEX43bdcZGdM5h4E2zP,n1r6XsVSD7eKrNWqpBsqXAVq3ZhVu8pWQ2m,1,1000000",
    "n1YRGnHTp3k7bFGfu8tbB2ucfbaui54ZymL,n1mg8VTQJZ4kGDnDrggQnJzn4v5bDpo8bzy,1,1000000",
    "n1NAUVx3LZGhgjVnPUMzdhMfknouBdMdTJP,n1nT22NaMhH9GfQCauKdXQ9CruQ72uuosmR,1,1000000",
    "n1FKncsEWdgyxA5rpoh5oKHpH43WECUE7L8,n1sNXeyJhsnknTfYiRQtFsS2epwUp3Ls9KH,1,1000000",
    "n1UQm7Gxpb6USxMxs3VuimML7XEke6xQ4ff,n1pohg1UWjFFEKb9XnqGRdscabcBrT7WbaF,1,1000000",
    "n1SzWZRdJcPWu4WiVScSqq7U5GHRwaRB72i,n1r6XsVSD7eKrNWqpBsqXAVq3ZhVu8pWQ2m,1,1000000"
  };
  // clang-format on

  for (size_t i = 0; i < sizeof(tx_hash_str) / sizeof(tx_hash_str[0]); i++) {
    auto tx_ptr =
        bab_ptr->get_transaction_api(neb::bytes::from_hex(tx_hash_str[i]));
    std::stringstream ss;
    ss << neb::string_to_byte(tx_ptr->from()).to_base58() << ','
       << neb::string_to_byte(tx_ptr->to()).to_base58() << ','
       << neb::storage_to_wei(neb::string_to_byte(tx_ptr->value())) << ','
       << neb::storage_to_wei(neb::string_to_byte(tx_ptr->gas_price()));
    EXPECT_EQ(ss.str(), transactions_str[i]);
  }
}

TEST(test_blockchain_api, get_transfer_event) {
  std::string db_path = get_db_path_for_read();

  auto rss_ptr = std::make_unique<neb::fs::rocksdb_session_storage>();
  rss_ptr->init(db_path, neb::fs::storage_open_default);

  auto bc_ptr = std::make_unique<neb::fs::blockchain>(rss_ptr.get());
  auto bab_ptr = std::make_unique<neb::fs::blockchain_api>(bc_ptr.get());

  auto block = bc_ptr->load_block_with_height(12800);

  std::string events_root_str = block->header().events_root();
  neb::bytes events_root_bytes = neb::string_to_byte(events_root_str);

  std::vector<neb::fs::transaction_info_t> ret;
  for (auto &tx : block->transactions()) {
    neb::fs::transaction_info_t info;
    std::string tx_hash_str = tx.hash();
    neb::bytes tx_hash_bytes = neb::string_to_byte(tx_hash_str);
    std::vector<neb::fs::transaction_info_t> events;
    bab_ptr->get_transfer_event(events_root_bytes, tx_hash_bytes, events, info);
    ret.push_back(info);
    if (info.m_status) {
      for (auto &e : events) {
        e.m_status = info.m_status;
        ret.push_back(e);
      }
    }
  }

  // clang-format off
  std::string transactions_str[] = {
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604",
    "1,23553",
    "1,22604"
  };
  // clang-format on

  for (size_t i = 0; i < ret.size(); i++) {
    auto &tx = ret[i];
    std::stringstream ss;
    ss << tx.m_status << ',' << tx.m_gas_used;
    EXPECT_EQ(ss.str(), transactions_str[i]);
  }
}
