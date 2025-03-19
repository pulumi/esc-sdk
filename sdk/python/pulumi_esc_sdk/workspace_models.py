# Copyright 2025, Pulumi Corporation.  All rights reserved.

"""
Models for Pulumi workspace and account logic for python SDK.
This is a partial port of ESC and Pulumi CLI code found in
https://github.com/pulumi/esc/tree/main/cmd/esc/cli/workspace
"""

from dataclasses import dataclass, field
import json
from typing import Dict, List, Optional

@dataclass
class Account:
    accessToken: str
    username: str
    organizations: List[str]

    @classmethod
    def from_json(self, data: dict):
        return self(
            accessToken=data.get("accessToken"),
            username=data.get("username"),
            organizations=data.get("organizations", []),
        )

@dataclass
class Credentials:
    current: str
    accessTokens: Dict[str, str]
    accounts: Dict[str, Account]

    @classmethod
    def from_json(self, dataJson: str):
        data = json.loads(dataJson)
        accounts: Dict[str, Account] = {}
        accounts_dict = data.get("accounts")
        if accounts_dict:
            for account_name, account_data in accounts_dict.items():
                accounts[account_name] = Account.from_json(account_data)
        return self(
            current=data.get("current"),
            accessTokens=data.get("accessTokens"),
            accounts=accounts
        )
