// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
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
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/urfave/cli"
)

var (
	// ConfigFlag config file path
	ConfigFlag = cli.StringFlag{
		Name:        "config, c",
		Usage:       "load configuration from `FILE`",
		Value:       "conf/default/config.conf",
		Destination: &config,
	}

	// NetworkSeedFlag network seed
	NetworkSeedFlag = cli.StringSliceFlag{
		Name:  "network.seed",
		Usage: "network seed addresses, multi-value support.",
	}

	// NetworkListenFlag network listen
	NetworkListenFlag = cli.StringSliceFlag{
		Name:  "network.listen",
		Usage: "network listen addresses, multi-value support.",
	}

	// NetworkKeyPathFlag network key
	NetworkKeyPathFlag = cli.StringFlag{
		Name:  "network.key",
		Usage: "network private key file path",
	}

	// NetworkFlags config list
	NetworkFlags = []cli.Flag{
		NetworkSeedFlag,
		NetworkListenFlag,
		NetworkKeyPathFlag,
	}

	// ChainIDFlag chain id
	ChainIDFlag = cli.UintFlag{
		Name:  "chain.id",
		Usage: "chain id",
	}

	// ChainDataDirFlag chain data dir
	ChainDataDirFlag = cli.StringFlag{
		Name:  "chain.datadir",
		Usage: "chain data storage dirctory",
	}

	// ChainKeyDirFlag chain key dir
	ChainKeyDirFlag = cli.StringFlag{
		Name:  "chain.keydir",
		Usage: "chain key storage dirctory",
	}

	// ChainCoinbaseFlag chain coinbase
	ChainCoinbaseFlag = cli.StringFlag{
		Name:  "chain.coinbase",
		Usage: "chain coinbase dirctory",
	}

	// ChainCipherFlag chain cipher
	ChainCipherFlag = cli.StringSliceFlag{
		Name:  "chain.ciphers",
		Usage: "chain signature ciphers, multi-value support.",
	}

	// ChainFlags chain config list
	ChainFlags = []cli.Flag{
		ChainIDFlag,
		ChainDataDirFlag,
		ChainKeyDirFlag,
		ChainCoinbaseFlag,
		ChainCipherFlag,
	}

	// RPCListenFlag rpc listen
	RPCListenFlag = cli.StringSliceFlag{
		Name:  "rpc.listen",
		Usage: "rpc listen addresses, multi-value support.",
	}

	// RPCHTTPFlag rpc http listen
	RPCHTTPFlag = cli.StringSliceFlag{
		Name:  "rpc.http",
		Usage: "rpc http listen addresses, multi-value support.",
	}

	// RPCModuleFlag rpc http module
	RPCModuleFlag = cli.StringSliceFlag{
		Name:  "rpc.module",
		Usage: "rpc support modules, multi-value support.",
	}

	// RPCFlags rpc config list
	RPCFlags = []cli.Flag{
		RPCListenFlag,
		RPCHTTPFlag,
		RPCModuleFlag,
	}

	// StatsEnableFlag stats enable
	StatsEnableFlag = cli.BoolFlag{
		Name:  "stats.enable",
		Usage: "stats enable metrics",
	}

	// StatsDBHostFlag stats db host
	StatsDBHostFlag = cli.StringFlag{
		Name:  "stats.dbhost",
		Usage: "stats influxdb host",
	}

	// StatsDBNameFlag stats db name
	StatsDBNameFlag = cli.StringFlag{
		Name:  "stats.dbname",
		Usage: "stats influxdb db name",
	}

	// StatsDBUserFlag stats db user
	StatsDBUserFlag = cli.StringFlag{
		Name:  "stats.dbuser",
		Usage: "stats influxdb user",
	}

	// StatsDBPasswordFlag stats db password
	StatsDBPasswordFlag = cli.StringFlag{
		Name:  "stats.dbpassword",
		Usage: "stats influxdb password",
	}

	// StatsFlags stats config list
	StatsFlags = []cli.Flag{
		StatsEnableFlag,
		StatsDBHostFlag,
		StatsDBNameFlag,
		StatsDBUserFlag,
		StatsDBPasswordFlag,
	}
)

func networkConfig(ctx *cli.Context, cfg *nebletpb.NetworkConfig) {
	if ctx.GlobalIsSet(NetworkSeedFlag.Name) {
		cfg.Seed = ctx.GlobalStringSlice(NetworkSeedFlag.Name)
	}
	if ctx.GlobalIsSet(NetworkListenFlag.Name) {
		cfg.Listen = ctx.GlobalStringSlice(NetworkListenFlag.Name)
	}
	if ctx.GlobalIsSet(NetworkKeyPathFlag.Name) {
		cfg.PrivateKey = ctx.GlobalString(NetworkKeyPathFlag.Name)
	}
}

func chainConfig(ctx *cli.Context, cfg *nebletpb.ChainConfig) {
	if ctx.GlobalIsSet(ChainIDFlag.Name) {
		cfg.ChainId = uint32(ctx.GlobalUint(ChainIDFlag.Name))
	}
	if ctx.GlobalIsSet(ChainDataDirFlag.Name) {
		cfg.Datadir = ctx.GlobalString(ChainDataDirFlag.Name)
	}
	if ctx.GlobalIsSet(ChainKeyDirFlag.Name) {
		cfg.Keydir = ctx.GlobalString(ChainKeyDirFlag.Name)
	}
	if ctx.GlobalIsSet(ChainCoinbaseFlag.Name) {
		cfg.Coinbase = ctx.GlobalString(ChainCoinbaseFlag.Name)
	}
	if ctx.GlobalIsSet(ChainCipherFlag.Name) {
		cfg.SignatureCiphers = ctx.GlobalStringSlice(ChainCipherFlag.Name)
	}
}

func rpcConfig(ctx *cli.Context, cfg *nebletpb.RPCConfig) {
	if ctx.GlobalIsSet(RPCListenFlag.Name) {
		cfg.RpcListen = ctx.GlobalStringSlice(RPCListenFlag.Name)
	}
	if ctx.GlobalIsSet(RPCHTTPFlag.Name) {
		cfg.HttpListen = ctx.GlobalStringSlice(RPCHTTPFlag.Name)
	}
	if ctx.GlobalIsSet(RPCModuleFlag.Name) {
		cfg.HttpModule = ctx.GlobalStringSlice(RPCModuleFlag.Name)
	}
}

func statsConfig(ctx *cli.Context, cfg *nebletpb.StatsConfig) {
	if ctx.GlobalIsSet(StatsEnableFlag.Name) {
		cfg.EnableMetrics = ctx.GlobalBool(StatsEnableFlag.Name)
	}
	if ctx.GlobalIsSet(StatsDBHostFlag.Name) {
		cfg.Influxdb.Host = ctx.GlobalString(StatsDBHostFlag.Name)
	}
	if ctx.GlobalIsSet(StatsDBNameFlag.Name) {
		cfg.Influxdb.Db = ctx.GlobalString(StatsDBNameFlag.Name)
	}
	if ctx.GlobalIsSet(StatsDBUserFlag.Name) {
		cfg.Influxdb.User = ctx.GlobalString(StatsDBUserFlag.Name)
	}
	if ctx.GlobalIsSet(StatsDBPasswordFlag.Name) {
		cfg.Influxdb.Password = ctx.GlobalString(StatsDBPasswordFlag.Name)
	}
}

// MergeFlags sets the global flag from a local flag when it's set.
func MergeFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}
