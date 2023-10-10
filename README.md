# chat-be
chat backend using redis and websocket


## 1 on 1 Chat Flow 
![one-on-one chat flow](https://res.cloudinary.com/dex4u3rw4/image/upload/v1696433470/untitled_1_kii12d.png)

-

![one-on-one chat flow 2](https://res.cloudinary.com/dex4u3rw4/image/upload/v1696433562/one-on-one_copy_v2my4l.jpg)

## Message table for 1 on 1 chat
![message table for 1 on 1 chat](https://res.cloudinary.com/dex4u3rw4/image/upload/v1696433591/figure-12-9-356WMC2A_o15qjo.webp)

- message_id menggunakan https://github.com/sony/sonyflake , shgg sudah terurut berdasarkan time. New rows memiliki ID yang lebih besar dibandingkan rows yg lama.
- 

## Quick Start
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

3. install docker
```
    windows: https://www.youtube.com/watch?v=XgRGI0Pw2mM
    mac: https://www.youtube.com/watch?v=-y1BmDbcaEU
        
```

4. buat container postgresql, redis di docker
```
    docker compose up -d
```

5. create migration file
```
    migrate create -ext sql -dir migrations create_chat_app_table
```

6. open pg-admin in localhost:5050. create database chat
7. execute this command in query-tool pg-admin
```
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

```
8. menjalankan migration
```
    migrate -database "postgres://user:pass@localhost:5431/chat?sslmode=disable" -path migrations up
```


### ref
- code template from https://github.com/evrone/go-clean-template
- image: https://bytebytego.com/courses/system-design-interview/design-a-chat-system