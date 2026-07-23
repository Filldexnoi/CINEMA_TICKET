# Cinema Ticket Booking System

ระบบจองตั๋วหนังแบบ Real-time ที่รองรับการแย่งที่นั่งพร้อมกันหลายคน (concurrency) โดยไม่ให้เกิดการจองซ้ำ (double booking)

---

## 1. System Architecture Diagram
![alt text](image.png)

ฝั่ง Frontend คุยกับ Backend ผ่าน Nginx ทั้ง REST API, OAuth redirect และ WebSocket ในคอนเทนเนอร์เดียว ส่วน Backend ภายในเป็น HTTP handler, WebSocket hub และ Kafka producer/consumer (รันเป็น goroutine คู่ขนานกัน) 

---

## 2. Tech Stack Overview

| Layer | เทคโนโลยีที่ใช้ |
|---|---|
| Backend | Go 1.25 + Gin |
| Frontend | Vue 3 |
| Database | MongoDB |
| Cache / Distributed Lock | Redis | 
| Message Queue | Apache Kafka |
| Realtime | WebSocket | 
| Authentication | Google OAuth 2.0 + JWT |
| Deployment | Docker + docker-compose |

---

## 3. Booking Flow (ทีละขั้นตอน)

1. **Login** — ผู้ใช้กด "Continue with Google" → redirect ไป Google consent screen → login สำเร็จ Google จะ redirect กลับมาที่ backend (`/auth/google/callback`) → backend แลก code เป็น profile, สร้าง/อัปเดต user ใน MongoDB, ออก JWT ของตัวเอง แล้ว redirect กลับไปหน้าเว็บพร้อม token
2. **เลือกหนัง → เลือกรอบฉาย** → เข้าหน้าผังที่นั่งซึ่งจะ fetch ผังที่นั่งปัจจุบันผ่าน REST ก่อน แล้วค่อยเปิด WebSocket ต่อเพื่อรับการอัปเดตแบบ real-time
3. **เลือกที่นั่ง** — ตอนนี้เป็นแค่การเลือกฝั่ง Frontend เท่านั้น (ยังไม่ได้ล็อกจริงที่ backend)
4. **กดปุ่ม "Continue"** — backend จะทำสองอย่างต่อกันในคำขอเดียว:
   - ขอ **Redis Distributed Lock** ให้ทุกที่นั่งที่เลือก (ล็อก 5 นาที ตาม `LOCK_TTL_SECONDS`)
   - ถ้าล็อกสำเร็จทุกที่นั่ง → สร้าง booking สถานะ `PENDING` ใน MongoDB ทันที แล้ว redirect ไปหน้าชำระเงินเลย 
   - ถ้ามีที่นั่งไหนถูกคนอื่นล็อกไปก่อน → คืน error, แสดง notification แจ้งผู้ใช้ และ**เคลียร์ที่เลือกทิ้งทั้งหมด**ให้เลือกใหม่
5. ทุกครั้งที่มีการล็อก/ปล่อยที่นั่ง backend จะ publish event เข้า Kafka แล้ว consumer จะ broadcast ผ่าน WebSocket ไปหาทุกคนที่เปิดหน้าผังที่นั่งรอบเดียวกันอยู่ — เห็นที่นั่งเปลี่ยนเป็น `LOCKED` แบบ real-time ภายในเวลาประมาณ 1 วินาที
6. หน้าชำระเงินจะมี **countdown timer** นับถอยหลังตามเวลาที่ล็อกเหลืออยู่ ผู้ใช้กด "Pay now" (จำลองการจ่ายเงิน ยังไม่ได้ต่อ payment gateway จริง)
7. **ถ้าจ่ายเงินสำเร็จ** — backend เช็ค Redis lock อีกรอบว่า token ยังตรงกับที่จองไว้ไหม (กันกรณี lock หมดอายุพอดีตอนกำลังจ่ายเงิน) → ถ้าผ่าน เปลี่ยนสถานะที่นั่งเป็น `BOOKED`, booking เป็น `CONFIRMED` → publish event `booking.confirmed` → ที่นั่งเปลี่ยนสีใน UI ทุกคนที่เปิดดูอยู่ทันที พร้อม trigger mock notification
8. **ถ้าไม่จ่ายเงินภายในเวลา** — lock หมดอายุ ที่นั่งจะถูกปลดล็อกกลับเป็น `AVAILABLE` เองอัตโนมัติ และ booking จะถูกเปลี่ยนเป็น `EXPIRED` 

---

## 4. Redis Lock Strategy
ใช้ Redis มีคำสั่ง `SET key value NX PX <ttl>` ที่เป็น atomic operation ในตัว — แปลว่าถ้ามี 2 คนกดเลือกที่นั่งเดียวกันพร้อมกันเป๊ะ Redis การันตีว่าจะมีแค่คนเดียวที่ `SET` สำเร็จ อีกคนจะได้ผลลัพธ์ว่าล้มเหลวทันทีและ TTL ในตัวยังช่วยให้ที่นั่งปลดล็อกอัตโนมัติเมื่อหมดเวลา

**รายละเอียดการทำงาน:**
- Key pattern: `lock:showtime:{showtimeID}:seat:{seatLabel}` (ล็อกแยกตามรอบฉาย เพราะที่นั่งเดียวกันคนละรอบฉายไม่เกี่ยวกัน)
- Value: `{userID}:{uuid}` — เป็น token เฉพาะของการล็อกครั้งนั้นๆ ใช้ตรวจสอบความเป็นเจ้าของภายหลัง
- คำสั่งล็อก: `SET key value NX PX 300000` 
- คำสั่งปลดล็อก ใช้ **Lua script** แบบ compare-and-delete (เช็คว่า value ตรงกับ token ของเราก่อนค่อยลบ) เพื่อป้องกันไม่ให้ไปลบล็อกของคนอื่นที่เพิ่งเข้ามาใหม่โดยบังเอิญ
- **MongoDB เก็บสถานะที่นั่ง (`AVAILABLE`/`LOCKED`/`BOOKED`) เป็นแค่ read-model สำหรับแสดงผลเท่านั้น** ตัวที่ตัดสินจริงว่า "ล็อกนี้ยังมีชีวิตอยู่ไหม" คือ Redis เสมอ

**การจัดการ lock หมดอายุ (ทำไว้ 3 ชั้น ป้องกันการตกหล่น):**
1. **Redis Keyspace Notification** — subscribe event `expired` จาก Redis โดยตรง ทำให้รู้ทันทีที่ล็อกหมดอายุ
2. **Lazy self-heal** — ทุกครั้งที่มีคน fetch ผังที่นั่ง ถ้าเจอที่นั่งที่ MongoDB ยังโชว์ `LOCKED` แต่เวลาหมดอายุผ่านไปแล้ว จะปลดล็อกให้ทันทีตรงนั้นเลย
3. **Sweeper (background job)** — รันทุกๆ ~20 วินาที ไล่หาที่นั่งที่หมดอายุแต่ยังไม่มีใคร trigger การปลดล็อก เป็นตัวสำรองตัวสุดท้ายเผื่อกรณี Redis notification หลุดหาย (เช่น Redis restart กลางคัน)

การมี 3 ชั้นนี้ทำให้มั่นใจว่าที่นั่งจะไม่ค้างสถานะ `LOCKED` ตลอดไปแม้ระบบใดระบบหนึ่งจะพลาดไป

---

## 5. Message Queue ใช้ทำอะไร

ใช้ Kafka เป็น **ตัวกลางระหว่างการเขียนข้อมูลกับการแจ้งเตือนแบบ real-time** โดยมี topic เดียวชื่อ `seat-events` เก็บ event ทั้งหมด: `seat.locked`, `seat.released`, `booking.confirmed`, `booking.expired`

**มี consumer 2 ตัว อ่าน topic เดียวกัน แต่ตั้งใจให้ใช้ consumer group ต่างกัน เพราะต้องการ semantic คนละแบบ:**

1. **WS Broadcaster** — อ่านแบบ **direct partition read (ไม่มี consumer group)**. ถ้ารัน backend หลาย instance พร้อมกัน แต่ละ instance จำเป็นต้อง**เห็น event ทุกตัว**เพื่อ broadcast ให้ client ที่ต่อ WebSocket เข้ามาที่ instance ของตัวเอง 
2. **Notification Consumer** — อ่านผ่าน **consumer group** (`notification-consumer`). งานนี้"แต่ละ event ประมวลผลแค่ครั้งเดียว" (ห้ามส่ง notification ซ้ำ) ถ้ารันหลาย instance โดยไม่มี group ทุก instance จะอ่าน event เดียวกันและส่ง notification ซ้ำกันหมด — consumer group ทำให้ Kafka รับประกันว่ามีแค่ instance เดียวในกลุ่มที่ประมวลผล event หนึ่งๆ (ปัจจุบันยังเป็น mock, log อย่างเดียว ยังไม่ได้ต่อ email/SMS จริง)

---

## 6. วิธีรันระบบ

### ขั้นตอน

```bash
# 1. คัดลอกไฟล์ env ตัวอย่าง
cp .env.example .env

# 2. สร้าง Google OAuth Client ที่ https://console.cloud.google.com/apis/credentials
#    Authorized redirect URI ต้องใส่ตรงตัวว่า:
#    http://localhost:8080/auth/google/callback
#    แล้วเอา Client ID / Client Secret มาใส่ในไฟล์ .env
#    (GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET)

# 3. รันทั้งระบบด้วยคำสั่งเดียว
docker compose up --build
```

จากนั้นเปิดเบราว์เซอร์ไปที่ **http://localhost:5173**

> ระบบจะสร้างข้อมูลตัวอย่างให้อัตโนมัติตอน backend เริ่มทำงานครั้งแรก (2 โรงหนัง, 3 เรื่อง, รอบฉายเรื่องละ 2 รอบ, ที่นั่งรอบละ 80 ที่)

### Port ทั้งหมดที่ใช้

| Service | Port | ใช้ทำอะไร |
|---|---|---|
| Frontend | 5173 | เว็บแอปหลัก |
| Backend | 8080 | REST API + WebSocket + OAuth callback |
| MongoDB | 27017 | ต่อด้วย mongosh/Compass เพื่อดู/แก้ข้อมูลตรงๆ |
| Redis | 6379 | ต่อด้วย redis-cli เพื่อ debug |
| Kafka | 9092 (ในเครือข่าย docker) / 9094 (จากเครื่อง host) | broker |
| Kafka UI | 8081 | เว็บดู topic/message ของ Kafka แบบ visual |

### Health checks

- `GET /health` — liveness, ตอบ `ok` เสมอถ้า process ยังรันอยู่ (ไม่เช็ค dependency)
- `GET /health/ready` — readiness, เช็ค connectivity ไปยัง MongoDB/Redis/Kafka จริง คืน `503` พร้อมระบุตัวที่พังถ้ามีอันไหนต่อไม่ได้ (ใช้เป็น readiness probe ถ้าจะรันบน k8s/orchestrator จริง)

### Unit Tests

Business logic ชั้น usecase (seat locking, booking, payment) และ RBAC middleware มี unit test แยกจาก integration/manual test — ใช้ in-memory fake implementation ของ `ports` interface ทั้งหมด รันได้เร็วโดยไม่ต้องพึ่ง Mongo/Redis/Kafka จริง:

```bash
cd backend
go test ./...
```

### API Collection

มี Postman collection พร้อมทดสอบทุก endpoint ไว้ที่ `postman_collection.json` (root ของโปรเจค) import เข้า Postman แล้วดูคำอธิบายวิธีใช้ auth ใน description ของ collection ได้เลย

### วิธีให้สิทธิ์ Admin

ระบบไม่มีหน้าสมัคร/จัดการ admin ต้อง promote user ให้เป็น admin เองผ่าน mongosh:

```bash
docker compose exec mongo mongosh cinema --eval 'db.users.updateOne({email:\"Your_Email\"}, {$set:{role:\"ADMIN\"}})''
```

---

## 7. Assumptions & Trade-offs

- **Role (USER/ADMIN) เก็บใน MongoDB ไม่ได้ฝังใน JWT** เพื่อให้การ promote เป็น admin มีผลทันทีโดยไม่ต้อง login ใหม่ ข้อเสียคือทุก request ที่ต้องเช็คสิทธิ์ admin จะมี query ไป MongoDB เพิ่ม 1 ครั้ง (ยอมรับได้เพราะเป็นแค่ index lookup ตรงๆ)
- **Kafka topic มี partition เดียว, consumer 2 ตัวใช้ consumer group ต่างกันตามความต้องการของงาน** — WS Broadcaster ไม่ใช้ consumer group (ต้องการให้ทุก instance เห็นทุก event เพื่อ broadcast ให้ client ของตัวเอง) ส่วน Notification Consumer ใช้ consumer group `notification-consumer` (ต้องการให้แต่ละ event ถูกประมวลผลแค่ครั้งเดียว กันส่ง notification ซ้ำถ้ามีหลาย instance)
- **Login รองรับเฉพาะ Google OAuth** ต้องตั้งค่า `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` ให้ถูกต้องก่อนถึงจะ login ได้จริง
- **JWT ส่งผ่าน query string ตอนต่อ WebSocket** (`/ws/showtimes/:id?token=...`) เพราะ browser ไม่สามารถแนบ custom header ตอนทำ WebSocket handshake ได้ ยอมรับ trade-off นี้เพราะ connection เป็น `wss/ws` ผ่าน HTTPS/reverse proxy อยู่แล้วในการใช้งานจริง

