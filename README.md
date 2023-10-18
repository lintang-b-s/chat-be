# chat-be
 chat backend api using websocket, redis, and postgres.

## ChatHub struct & User struct
![ChatHub struct & User Struct](https://res.cloudinary.com/dex4u3rw4/image/upload/v1697371903/ChatHubdanUser_vjzyku.png)



## 1 on 1 Chat Flow (UserA dan UserB terhubung ke instan chat-server yang sama)
![one-on-one chat flow server sama ](https://res.cloudinary.com/dex4u3rw4/image/upload/v1697371903/one-on-one-server-sama_ohixbk.png)

## 1 on 1 Chat Flow (Semua user terhubung ke instan chat-server yang berbeda)
![one-on-one chat flow server beda](https://res.cloudinary.com/dex4u3rw4/image/upload/v1697371913/one-on-one-beda-server_n2so7t.png)

-

![one-on-one chat flow 2](https://res.cloudinary.com/dex4u3rw4/image/upload/v1696433562/one-on-one_copy_v2my4l.jpg)

## Message table for 1 on 1 chat
![message table for 1 on 1 chat](https://res.cloudinary.com/di0pjroxh/image/upload/v1697274479/one-on-one-messageTable_ptublq.png)



## Quick Start
branch ini hanya bisa dijalankan di os linux/mac (karena memakai library https://github.com/cloudwego/netpoll)
.switch ke branch windows utk bisa jalankan di os manapun


1. install golang: 
```
 Windows: https://www.youtube.com/watch?v=xYpqI7GRrvE
 Mac: https://www.youtube.com/watch?v=HrFqH6Dj6kk
 Linux: https://www.youtube.com/watch?v=G36tXSWlUnE
 
 GOPATH: 
 https://www.youtube.com/watch?v=eJVq-idZDMo
 windows: https://www.youtube.com/watch?v=kjr3mOPv8Sk
 windows & unix: https://github.com/golang/go/wiki/SettingGOPATH
 mac: https://www.youtube.com/watch?v=FTDOW8UbKjQ
```


2. install go-migrate
```
    go install -tags "postgres,mysql" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    untuk mengecek sudah diinstal belum: ketik migrate
```

3. install makefile
```
    windowss: https://linuxhint.com/run-makefile-windows/
    mac: https://formulae.brew.sh/formula/make
    linux:  https://www.youtube.com/watch?v=PLFzCOPPrPc
    
```

3. install docker
```
    windows: https://www.youtube.com/watch?v=XgRGI0Pw2mM
    mac: https://www.youtube.com/watch?v=-y1BmDbcaEU
        
```

4. jalankan redis & postgresqql di local;
```
 redis:
  mac: brew services start redis
  windows: https://linuxhint.com/install-run-redis-windows/

  
```

4. (optional jika belum install redis & postgres) buat container postgresql, redis di docker
```
    docker compose up -d
```

5. create migration file
```
    migrate create -ext sql -dir migrations create_chat_app_table
```

6. open pg-admin in localhost:5050 or enter postggres cli. create database chat
7. execute this command in query-tool pg-admin
```
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

```
8. menjalankan migration
```
    migrate -database "postgres://postgres:pass@localhost:5432/chat?sslmode=disable" -path migrations up
```

9. buat file .env dan isi sesuai .env.example

10. membuat dokumentasi swaggger
```
    make swag-v1
```

11. menjalankan aplikasi
```
    make run
```

12. import postman collection di docs/pelatihan umum(chat app)postman_collection.json & jalankan request di postman

13. buka collection https://app.getpostman.com/join-team?invite_code=3aa872d85da6ae474265256597513a0a&target_code=64d73c34516d85e5cb485c9c40a91be8
 & ganti query parameter otp dan username dg otp dan username yang diberikan saat login 
```
    jalankan request websocket di link postman tsb
```


## Online Presence

### Login & Logout
![user login mechanism](https://res.cloudinary.com/dex4u3rw4/image/upload/v1697362511/UserLoginLogout_cy1qtu.png)


### Heartbeat
![heartbeat mechanism](https://res.cloudinary.com/di0pjroxh/image/upload/v1697272518/untitled_2_prfsgc.png)

## Online Status Fanout

### Flow
![online status fanout](https://res.cloudinary.com/dex4u3rw4/image/upload/v1697362511/online-status-fanout_ymceo7.png)

### Online status fanout message
![online status message fanout](https://res.cloudinary.com/di0pjroxh/image/upload/v1697272778/online-fanout-message_fb6q0e.png)


## Kenapa menggunakan redis pubsub untuk menerima chat dari user lain?
Redis' Pub/Sub exhibits at-most-once message delivery semantics. As the name suggests, it means that a message will be delivered once if at all. Once the message is sent by the Redis server, there's no chance of it being sent again. If the subscriber is unable to handle the message (for example, due to an error or a network disconnect) the message is forever lost.
<br/><br/>
Karena jika subscriber offline subscriber, lalu online lagi. user subsriber tidak akan mendapatkan message yg dikrimkan user lain saat dia offline, melainkan hanya mendapatkan
message dari pubSub  saat dia online dan user lain mengirimkan message. Message yg dikrimkan user lain saat user offline akan difetch dari database postgresql


### ref
 slides: https://docs.google.com/presentation/d/1YfvlHW2Rf2RRQYP2sZ-hJteuNslNzviLhz0bMs3jT0M/edit?usp=sharing
- code template from https://github.com/evrone/go-clean-template
- image: https://bytebytego.com/courses/system-design-interview/design-a-chat-system
