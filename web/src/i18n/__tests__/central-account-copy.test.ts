/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { expect, it } from 'vitest'

import zhTW from '../locales/zh-TW.json'
import zhCN from '../locales/zh.json'

const button = 'Sign in with your RoboCoding account'
const description =
  'Use one RoboCoding account to sign in securely across RoboCoding services.'

it('keeps central-account sign-in copy inside the Simplified Chinese namespace', () => {
  expect(zhCN.translation[button]).toBe('使用 RoboCoding 账号登录')
  expect(zhCN.translation[description]).toBe(
    '使用同一个 RoboCoding 账号安全登录各项 RoboCoding 服务。'
  )
})

it('keeps central-account sign-in copy inside the Traditional Chinese namespace', () => {
  expect(zhTW.translation[button]).toBe('使用 RoboCoding 帳號登入')
  expect(zhTW.translation[description]).toBe(
    '使用同一個 RoboCoding 帳號安全登入各項 RoboCoding 服務。'
  )
})
