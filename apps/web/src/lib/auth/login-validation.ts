export function validateLogin(email: string, password: string) {
  const normalizedEmail = email.trim().toLowerCase();
  if (!normalizedEmail) return { email: "Email wajib diisi" };
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalizedEmail))
    return { email: "Masukkan alamat email yang valid" };
  if (!password) return { password: "Password wajib diisi" };
  return {};
}
