# chat-be
 highly scalable chat backend api using websocket, redis, and postgres.
 <br/> <br/>
 fitur: 
```
 1. autentikasi (login, register, logout, renewAccToken)
 2. add user to contact & get user contact
 3. private chat dg chatbot
 4. private_chat dg user lain yang sudah ditambahkan ke contact
 5. create group chat , add member to group chat, remove member from group chat
 6. group chat dg user yg sudah ditambahkan sebagai member groupchat
 7. memanggil chatbot di groupchat 
 8. mendapatkan semua message private chat & message private by friend
 9. mendapatkan semua message di groupchat
 10. Meski user terhubung ke chat-server yang berbeda-beda , mereka tetap akan bisa berkomunikasi karena adanya redis pub-sub 
```

## Content
- [General System Design Part 1](##1-on-1-chat-flow-usera-dan-userb-terhubung-ke-instan-chat-server-yang-sama)
- [API Docs](#api-docs)
- [Quick Start / Cara menjalankan Aplikasi](#quick-start)
- [General System Design Part 2](#online-presence)


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

Messages_id digenerate menggunakan https://github.com/sony/sonyflake, yg mana besarnya integer id yg digenerate tergantung waktu saat id digeneraete
<br/>
untuk mendapatkan messages yang diurutkan berdasarkan time (dari paling lama -> terbaru) bisa pakai query "order by id asc".

## Group Chat Flow
![group-chat flow](https://bytebytego.com/_next/image?url=%2Fimages%2Fcourses%2Fsystem-design-interview%2Fdesign-a-chat-system%2Ffigure-12-14-DRZR5QM7.png&w=1080&q=75)

## Message table for group chat
![group-chat message table](https://bytebytego.com/_next/image?url=%2Fimages%2Fcourses%2Fsystem-design-interview%2Fdesign-a-chat-system%2Ffigure-12-10-2TIQVS3D.png&w=750&q=75)

## API docs
ada di docs/swagger.yaml <br/> <br/>
postman collection ada di https://app.getpostman.com/join-team?invite_code=3aa872d85da6ae474265256597513a0a&target_code=64d73c34516d85e5cb485c9c40a91be8 & jalankan di postman dekstop karena localhost ( setiap koneksi websocket harus menggunakan otp / harus login lagi)

## Quick Start


1. install golang & set GOPATH Variabel di OS kalian: 
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

9. buat file .env dan isi sesuai .env.example & buat akun https://www.edenai.co/ & copy EDENAPI_KEY ke .env dg key "EDENAI_APIKEY"
10. ubah config.yml sesuai addr,username, password postgres dan redis anda

11. membuat dokumentasi swaggger
```
    make swag-v1
```

12. menjalankan aplikasi
```
    make run
```

13. jalankan request postman collection yg lengkap & ada websocket di https://app.getpostman.com/join-team?invite_code=3aa872d85da6ae474265256597513a0a&target_code=64d73c34516d85e5cb485c9c40a91be8 & jalankan di postman dekstop karena localhost atau  import postman collection di docs/pelatihan umum(chat app)postman_collection.json (tapi tidak lengkap & tidak ada websocket) & jalankan request di postman 

14. buka collection https://app.getpostman.com/join-team?invite_code=3aa872d85da6ae474265256597513a0a&target_code=64d73c34516d85e5cb485c9c40a91be8
 & ganti query parameter otp dan username dg otp dan username yang diberikan saat login 
```
    jalankan request websocket di link postman tsb
    setiap koneksi websocket harus menggunakan otp (harus login lagi)
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
