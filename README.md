# VK Check Online Bot
A bot for Telegram that allows you to notify you that the person you selected from VK.com online or offline

## Configuration

Change variables in file shared/const.go:
- *Token* - token for telegram bot
- *AccessTokenVk* - token for access VK.com

## Build

For bot application
```
cd bot/main
go build
go install
```
For notification application
```
cd bot/notification
go build
go install
```
