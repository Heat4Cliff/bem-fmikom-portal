# Prompt Templates for Sequential Codex Runs

## Continue next phase
```text
Lanjutkan project BEM FMIKOM UNUGHA Portal dengan mengikuti AGENTS.md dan PLANS.md.
Baca report phase sebelumnya terlebih dahulu. Kerjakan PHASE <N> SAJA. Jangan mulai phase berikutnya.
Jalankan seluruh quality gate yang relevan, update PLANS.md hanya untuk item yang benar-benar terverifikasi, lalu buat Phase <N> Report menggunakan template docs/codex/PHASE_REPORT_TEMPLATE.md.
Jika ada visual screen yang disentuh, wajib bandingkan dengan Stitch screen.png terkait. Jika browser/manual verification belum dilakukan, tulis MANUAL VERIFICATION REQUIRED.
```

## Fix a defect
```text
Perbaiki defect berikut tanpa memperluas scope: <DEFECT>.
Sebelum mengubah code, reproduksi dan catat expected vs actual. Cari root cause, buat regression test bila memungkinkan, lakukan perbaikan terkecil yang maintainable, lalu jalankan quality gate terkait. Jangan menyatakan PASS hanya dari lint/typecheck bila defect bersifat interaksi/visual.
```

## Visual compliance run
```text
Audit visual route <ROUTE> terhadap design_reference/.../<SCREEN>/screen.png. Jangan redesign. Identifikasi perbedaan typography, spacing, color, responsive layout, states, dan component density. Perbaiki hanya deviasi yang tidak disebabkan oleh data dinamis. Catat hasil sebagai VISUAL REVIEW dan tandai MANUAL VERIFICATION REQUIRED bila belum divalidasi di browser pada viewport target.
```
