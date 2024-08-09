package keeper_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/mint/keeper"
	minttestutil "cosmossdk.io/x/mint/testutil"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var minterAcc = authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)

type GenesisTestSuite struct {
	suite.Suite

	sdkCtx        context.Context
	keeper        keeper.Keeper
	cdc           codec.BinaryCodec
	accountKeeper types.AccountKeeper
	env           appmodule.Environment
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	testCtx := coretesting.Context()
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, mint.AppModule{})

	// gomock initializations
	ctrl := gomock.NewController(s.T())
	s.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	s.sdkCtx = testCtx

	stakingKeeper := minttestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := minttestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := minttestutil.NewMockBankKeeper(ctrl)
	s.accountKeeper = accountKeeper
	accountKeeper.EXPECT().GetModuleAddress(minterAcc.Name).Return(minterAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAccount(s.sdkCtx, minterAcc.Name).Return(minterAcc)

	env := runtime.NewEnvironment(coretesting.KVStoreService(s.sdkCtx, types.StoreKey), log.NewNopLogger())
	s.keeper = keeper.NewKeeper(s.cdc, env, stakingKeeper, accountKeeper, bankKeeper, "", "")
	s.env = env
}

func (s *GenesisTestSuite) TestImportExportGenesis() {
	genesisState := types.DefaultGenesisState()
	genesisState.Minter = types.NewMinter(math.LegacyNewDecWithPrec(20, 2), math.LegacyNewDec(1))
	genesisState.Params = types.NewParams(
		"testDenom",
		math.LegacyNewDecWithPrec(15, 2),
		math.LegacyNewDecWithPrec(22, 2),
		math.LegacyNewDecWithPrec(9, 2),
		math.LegacyNewDecWithPrec(69, 2),
		uint64(60*60*8766/5),
		math.ZeroInt(),
	)

	err := s.keeper.InitGenesis(s.sdkCtx, s.accountKeeper, genesisState)
	s.Require().NoError(err)

	minter, err := s.keeper.Minter.Get(s.sdkCtx)
	s.Require().Equal(genesisState.Minter, minter)
	s.Require().NoError(err)

	invalidCtx := coretesting.Context()
	coretesting.KVStoreService(invalidCtx, types.StoreKey)
	_, err = s.keeper.Minter.Get(invalidCtx)
	s.Require().ErrorIs(err, collections.ErrNotFound)

	params, err := s.keeper.Params.Get(s.sdkCtx)
	s.Require().Equal(genesisState.Params, params)
	s.Require().NoError(err)

	genesisState2, err := s.keeper.ExportGenesis(s.sdkCtx)
	s.Require().NoError(err)
	s.Require().Equal(genesisState, genesisState2)
}
