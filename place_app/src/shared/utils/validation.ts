const PHONE_REGEX = /^1(3\d|4[5-9]|5[0-35-9]|6[6]|7[0-8]|8\d|9[0-35-9])\d{8}$/;

export function validatePhone(phone: string): string | null {
  if (!phone || phone.length !== 11) return '请输入 11 位手机号';
  if (!PHONE_REGEX.test(phone)) return '请输入正确的手机号';
  return null;
}

export function checkPwdStrength(pwd: string) {
  if (!pwd) return { score: 0, hint: '' };
  const hasLetter = /[a-zA-Z]/.test(pwd);
  const hasDigit = /\d/.test(pwd);
  const hasSpecial = /[^a-zA-Z0-9]/.test(pwd);
  const score = +hasLetter + +hasDigit + +hasSpecial;
  let hint = '';
  if (score === 1) hint = '💡 密码建议使用字母、数字、符号中的至少两种';
  else if (score === 2) hint = '✅ 密码安全性良好';
  else if (score === 3) hint = '✅ 密码安全性高';
  return { score, hint, hasLetter, hasDigit, hasSpecial };
}

export function maskPhone(phone: string): string {
  if (phone.length !== 11) return phone;
  return `${phone.slice(0, 3)} **** ${phone.slice(7)}`;
}
