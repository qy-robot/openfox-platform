/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  Bot,
  ChevronRight,
  Layers3,
  ScanLine,
  ShieldCheck,
  Sparkles,
  Waypoints,
  Workflow,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

const roboticsCards = [
  { title: '移动底盘', description: '路径规划与实时控制', icon: Waypoints },
  { title: '机械臂协作', description: '动作编排与安全边界', icon: Bot },
  { title: '视觉巡检', description: '识别、定位与任务回传', icon: ScanLine },
]

const capabilities = [
  { number: '01', icon: Layers3 },
  { number: '02', icon: Workflow },
  { number: '03', icon: ShieldCheck },
]

const capabilityCopy = [
  ['模型统一接入', '一个接口接入常用模型。'],
  ['技能随项目使用', '需要什么技能，就在项目里调用什么。'],
  ['用量权限清楚', '请求、权限和花费，一眼能看见。'],
]

export function ProductHome(props: { isAuthenticated: boolean }) {
  const { i18n, t } = useTranslation()
  const consolePath = props.isAuthenticated ? '/dashboard' : '/sign-in'
  const isCjk = /^(zh|ja|ko)/i.test(i18n.language)

  return (
    <main className={`openfox-home${isCjk ? ' openfox-home--cjk' : ''}`}>
      <section className='openfox-hero'>
        <div className='openfox-hero-shell'>
          <div className='openfox-hero-copy'>
            <h1>{t('把模型接好，')}<em>{t('把事情做完。')}</em></h1>
            <p className='openfox-hero-lede'>{t('把代码、模型、机器人技能和项目放在一起，专注把机器人跑起来。')}</p>
            <div className='openfox-hero-actions'><Button className='openfox-button-primary' render={<Link to={consolePath} />}>{t('进入工作台')}<ArrowRight aria-hidden='true' /></Button><Button variant='outline' className='openfox-button-secondary' render={<Link to='/pricing' />}>{t('查看方案')}<ChevronRight aria-hidden='true' /></Button></div>
          </div>
          <div className='openfox-model-shelf' aria-label={t('机器人项目预览')}>
            <div className='openfox-model-cards'>{roboticsCards.map(({ title, description, icon: Icon }, index) => <article key={title} className={`openfox-model-card${index === 0 ? ' is-featured' : ''}`}><span className={`openfox-model-icon model-dot-${index}`} aria-hidden='true'><Icon /></span><h3>{t(title)}</h3><p>{t(description)}</p></article>)}</div>
          </div>
        </div>
        <div className='openfox-hero-grid' aria-hidden='true' />
      </section>

      <section className='openfox-proof'><div className='openfox-proof-inner'><span>{t('适合')}</span><div className='openfox-proof-logos'><b>{t('产品团队')}</b><b>{t('机器人实验室')}</b><b>{t('AI 开发者')}</b><b>{t('平台工程师')}</b></div></div></section>

      <section className='openfox-capabilities' aria-labelledby='openfox-capabilities-title'><div className='openfox-section-intro'><h2 id='openfox-capabilities-title'>{t('少一点折腾，')}<br /><em>{t('多一点进展。')}</em></h2><p>{t('把模型、技能和协作放在同一个地方。')}</p></div><div className='openfox-capability-list'>{capabilities.map(({ number, icon: Icon }, index) => <article key={number} className='openfox-capability'><span className='openfox-capability-number'>{number}</span><div className='openfox-capability-icon'><Icon aria-hidden='true' /></div><div><h3>{t(capabilityCopy[index][0])}</h3><p>{t(capabilityCopy[index][1])}</p></div><ArrowRight aria-hidden='true' /></article>)}</div></section>


      <section className='openfox-cta'><div className='openfox-cta-mark'><Sparkles aria-hidden='true' /></div><h2>{t('现在就开始。')}</h2><p>{t('把手上的项目接进来，先跑起来。')}</p><Button className='openfox-button-primary' render={<Link to={consolePath} />}>{t('进入 OpenFox')} <ArrowRight aria-hidden='true' /></Button></section>
    </main>
  )
}
