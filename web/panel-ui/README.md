# SGE Web UI Panel 🎨

S.A.K.I.N. projesi için modern, yüksek performanslı ve estetik yönetim paneli.

## Özellikler

- **Modern Teknoloji Yığını:** Next.js 14, React, Tailwind CSS.
- **Premium Tasarım:**
    - **Shadcn UI** bileşen kütüphanesi.
    - **Glassmorphism** efektleri ve mesh gradient arkaplanlar.
    - **Inter & JetBrains Mono** fontları.
- **Fonksiyonalite:**
    - **Gerçek Zamanlı Veri:** Backend API'den (`localhost:8080`) canlı istatistik takibi.
    - **Demo Modu:** Backend kapalıyken bile arayüzü test etmek için tek tuşla simülasyon modu.
    - **Global Sidebar:** Sayfalar arası kalıcı navigasyon.

## Kurulum ve Çalıştırma

Bağımlılıkları yükleyin:
```bash
npm install
```

Geliştirme sunucusunu başlatın:
```bash
npm run dev
```

Panel `http://localhost:3000` adresinde çalışacaktır.

## Dizin Yapısı

- `app/` - Next.js App Router sayfaları.
    - `layout.tsx` - Global layout (Sidebar, Header).
    - `page.tsx` - Dashboard ana sayfası.
- `components/` - Yeniden kullanılabilir UI bileşenleri.
    - `ui/` - Shadcn temel bileşenleri (Card, Button vb.).
    - `dashboard/` - Dashboard'a özel widgetlar (Charts, Stats).
    - `layout/` - Sidebar ve Header bileşenleri.
- `lib/` - Yardımcı fonksiyonlar (`utils.ts`).
