# symblon

### Domain analysis

Result of the domain analysis it to provide a clear understanding of this domain and its components. Symblon is a system that issues a symbol to a user, user collects symbols and builds a public profile of his acheivements. Symblon should give cudos but also celebrate the mistakes during that process. Has two main billing models:
- **Personal** - user pays for the service
- **Organization** - organization pays for the service

Has a support for theaming, like:
- **Scouts** - badges
- **Wizards** - spells
- **Hackers** - hacks

Main concepts are:
- **Symbol** - a symbol that user can collect
- **User** - a user of the system
- **Symbol manager** - a person who manages symbols, can be a user or an organization
- **Symbol agent** - a person or a program that issues symbols to users based on some criteria

Symbols are split into two types:
- **Real time** - symbols that are issued in real time, thay have enough information to be issued immediately
- **Temporal** - symbols that are issued when some conditions are met, they must have a time window and a quantifier

Symbols are issued per an organization, or per a user. This is a hard dependency tied directly to the gihub organizational strategy. Symbols can have a multiplier, and a limit.
Temporal symbols evaluation is always done by organization (even if it is an user), and a user.

Temporal agents are used to evaluate temporal symbols, evaluation is triggered when a new user action is detected. Temporal agents are defined as a yaml file, and can be registred in the system, every user can register a temporal agent, and those can also be billed.
There could be a whole ecosystem of temporal agents, so developers can create their own agents and register them in the system.

```yaml
agent:
    type: temporal
    name: "My Temporal Agent"
    description: "This agent evaluates temporal symbols based on user actions."
    rules:
        - achievment: "Bug squasher"
          description: "Squash 5 bugs in 48h in the system to earn this symbol."
          symbol: "/my-temporal-agent/bug-squasher"

```
