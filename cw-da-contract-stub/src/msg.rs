use std::collections::HashSet;

use cosmwasm_schema::{cw_serde, QueryResponses};

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Config)]
    Config {},
    #[returns(BlockVotesResponse)]
    BlockVotes { height: u64, hash: String },
}

#[cw_serde]
pub enum ExecuteMsg {}

#[cw_serde]
pub struct Config {
    pub consumer_id: String,
    pub activated_height: u64,
}

#[cw_serde]
pub struct BlockVotesResponse {
    pub fp_pubkey_hex_list: HashSet<String>,
}
