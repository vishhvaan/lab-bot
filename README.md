# ::: Lab Bot :::

The lab bot is a _fast_ program that helps labs automate tasks.
It operates on Slack (and maybe other messaging platforms in the future) and can respond to messages, automatically perform tasks, and monitor channels.

Use cases:
- Create lab meeting slide deck based on the subset of the lab who are presenting that week
- Keep track of journal club articles and maintain an evolving list of potential articles
- Respond with pdfs of articles given a DOI (served from institution library resources)
- Implements OpenAI's GPT3 for responses to questions
  <!-- - Can be made to write papers in the future -->
- Turn on the lab's coffee machine or other devices (eg. hot water bath, PCR machine, etc.) remotely by messaging the bot
  - Can be made to automatically turn on/off these devices based on a schedule
- Send emails with the lab's email account with messages to the bot (approved by a user other than the one sending the message)
- Keep track of lab chores and provide reminders based on set schedules
- Keep track of birthdays and send a lab-wide message on the day
- Place orders for lab equipment over messages (implemented down the road)
  - Keeps a database of orders and create summaries of purchases and patterns
- Many others!

The bot is meant to supplement and enhance a lab's communication system. 
Users can directly translate communication into action by talking with the bot.
Labs can offload to the bot some tasks normally relegated to lab techs.
Plus it's fun to chat with a bot sometimes (especially when it can do simple or administrative things for you).

## Installation

Automatically built binaries of the bot for Mac, Windows, and Linux are coming soon!

In the meantime, build it: 
```
cd cmd/bot && go build
```
Building for embedded devices can be done by targeting an architecture. For instance, for a device with an ARM CPU: 
```
cd cmd/bot && GOOS=linux GOARCH=arm go build
```

## Usage

Order of your command fields matter, however, `@lab-bot` can be called anywhere in the message.

### Basic Commands

Customizable with this [file](slack/callbacks.go)

- `@lab-bot [hello/hai/hey/sup/hi]` : Greets the bot
- `@lab-bot [bye/goodbye/tata]` : Wishes farewell to the bot
- `@lab-bot thank you` : thanks the bot
- `@lab-bot sysinfo` : returns the system information - hostname, date/timezone, operating system, memory usage, uptime

### Controller Commands

Controllers are defined with keywords.
These keywords are used to interact with that specific job.
In these examples, we'll use the keyword `coffee` to represent a controller job which turns on/off the coffee machine.
This machine can be any device in the lab.

- `@lab-bot coffee` : Prints the status of the coffee controller - current power status, uptime, on/off schedules
- `@lab-bot coffee status` : Prints the status like above
- `@lab-bot coffee schedule status` : Prints the status like above
- `@lab-bot coffee [on/off]` : Turns on/off the machine
- `@lab-bot coffee schedule [on/off] set <cron>` : Schedules on/off jobs for the controller at specified times. Schedules use [cron syntax](https://en.wikipedia.org/wiki/Cron). On and off schedules are set independently. Examples of cron syntax are below.
- `@lab-bot coffee schedule [on/off] remove` : Removes the on/off scheduled job from the controller. On and off schedules are also removed independently.

```
Min  Hour Day  Mon  Weekday

*    *    *    *    *  

┬    ┬    ┬    ┬    ┬
│    │    │    │    └─  Weekday  (0=Sun .. 6=Sat)
│    │    │    └──────  Month    (1..12)
│    │    └───────────  Day      (1..31)
│    └────────────────  Hour     (0..23)
└─────────────────────  Minute   (0..59)

```
Examples:
```
0  *  * * * 	  : every hour
0  18 * * 0-6 	  : every week Mon-Sat at 6pm
10 2  * * 6,7     : every Sat and Sun on 2:10am

@lab-bot coffee schedule on set 0 8 * * 1-5   : turn on the coffee machine every weekday at 8am
```