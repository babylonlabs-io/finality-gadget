use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};

pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    Ok(Response::new())
}

pub fn query(_deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    use QueryMsg::*;

    match msg {
        QueryMsg::Config {} => Ok(to_json_binary(&query::query_config()?)?),
        BlockVotes { height, hash } => to_json_binary(&query::block_votes(height, hash)?),
    }
}

mod query {

    use std::collections::HashSet;

    use crate::msg::{BlockVotesResponse, Config};

    use super::*;

    pub fn query_config() -> StdResult<Config> {
        let config = Config {
            consumer_id: "op-stack-l2-12345".to_string(),
            activated_height: 1022293,
        };
        return Ok(config);
    }

    pub fn block_votes(height: u64, hash: String) -> StdResult<BlockVotesResponse> {
        if height > 2 {
            return Ok(BlockVotesResponse {
                fp_pubkey_hex_list: HashSet::new(),
            });
        }

        // Create a new HashSet
        let mut my_set: HashSet<String> = HashSet::new();

        if hash == "0x3aa074144a25d3ed71c7353a20c579650e0c56a993444c6156d44bb90b932f0d" {
            my_set.insert(String::from("pk1"));
            my_set.insert(String::from("pk2"));
            my_set.insert(String::from("pk3"));
        } else {
            my_set.insert(String::from("pk-that-voted-for-a-fork"));
        }

        Ok(BlockVotesResponse {
            fp_pubkey_hex_list: my_set,
        })
    }
}

pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    unimplemented!()
}
