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
*/
import { Link } from '@tanstack/react-router'
import {
  ArrowDownToLine,
  ArrowRight,
  ArrowUpRight,
  Check,
  Code2,
  Layers,
  Users,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

const platformHighlights = [
  {
    index: '01',
    title: 'Build from a real task',
    description:
      'Keep the brief, the files, and the next useful action in the same place.',
    icon: Code2,
  },
  {
    index: '02',
    title: 'Bring the right tools',
    description:
      'Add robotics skills and model services when the work calls for them.',
    icon: Layers,
  },
  {
    index: '03',
    title: 'Move with your team',
    description:
      'Share projects, usage, and decisions without adding another layer of process.',
    icon: Users,
  },
]

const workflowSteps = [
  [
    'Bring your project',
    'Open a local project and describe the outcome you want.',
  ],
  [
    'Choose your tools',
    'Connect the model service and skills that fit the job.',
  ],
  ['Review the work', 'See what changed, keep what works, and keep moving.'],
]

export function ProductHome(props: { isAuthenticated: boolean }) {
  const { i18n, t } = useTranslation()
  const consolePath = props.isAuthenticated ? '/dashboard' : '/sign-in'
  const isCjk = /^(zh|ja|ko)/i.test(i18n.language)

  return (
    <main className={`robo-home${isCjk ? ' robo-home--cjk' : ''}`}>
      <section className='robo-home-hero'>
        <div className='robo-home-hero-glow' aria-hidden='true' />
        <div className='robo-home-hero-inner'>
          <div className='robo-home-kicker'>
            <span className='robo-home-kicker-mark' aria-hidden='true' />
            {t('Your workspace for building robots')}
          </div>
          <h1>{t('Build things that work.')}</h1>
          <p className='robo-home-lede'>
            {t(
              'RoboCoding brings code, robotics skills, and AI services together so every task can become a working result.'
            )}
          </p>
          <div className='robo-home-actions'>
            <Button
              className='robo-home-primary-action'
              render={<Link to={consolePath} />}
            >
              {t('Open workspace')}
              <ArrowRight aria-hidden />
            </Button>
            <Button
              variant='outline'
              className='robo-home-secondary-action'
              render={<Link to='/download' />}
            >
              <ArrowDownToLine aria-hidden />
              {t('Download desktop app')}
            </Button>
          </div>
          <div className='robo-home-hero-note'>
            <span>{t('One account')}</span>
            <span aria-hidden='true'>·</span>
            <span>{t('Desktop + cloud')}</span>
            <span aria-hidden='true'>·</span>
            <span>{t('Built for real work')}</span>
          </div>
        </div>
        <div className='robo-home-orbit' aria-hidden='true'>
          <span className='robo-home-orbit-core' />
          <span className='robo-home-orbit-line robo-home-orbit-line-one' />
          <span className='robo-home-orbit-line robo-home-orbit-line-two' />
          <span className='robo-home-orbit-dot robo-home-orbit-dot-one' />
          <span className='robo-home-orbit-dot robo-home-orbit-dot-two' />
        </div>
      </section>

      <section
        className='robo-home-highlights'
        aria-labelledby='robo-highlights-title'
      >
        <div className='robo-home-section-heading'>
          <p className='robo-home-eyebrow'>{t('The way we work')}</p>
          <h2 id='robo-highlights-title'>
            {t('Less ceremony. More progress.')}
          </h2>
        </div>
        <div className='robo-home-highlight-grid'>
          {platformHighlights.map(
            ({ index, title, description, icon: Icon }) => (
              <article key={title} className='robo-home-highlight'>
                <div className='robo-home-highlight-topline'>
                  <span>{index}</span>
                  <Icon aria-hidden />
                </div>
                <h3>{t(title)}</h3>
                <p>{t(description)}</p>
                <span className='robo-home-highlight-arrow' aria-hidden='true'>
                  <ArrowUpRight />
                </span>
              </article>
            )
          )}
        </div>
      </section>

      <section
        className='robo-home-workflow'
        aria-labelledby='robo-workflow-title'
      >
        <div className='robo-home-workflow-intro'>
          <p className='robo-home-eyebrow'>{t('Built around the work')}</p>
          <h2 id='robo-workflow-title'>
            {t('From a clear idea to a useful result.')}
          </h2>
          <p>
            {t(
              'A focused workspace for people who want to make, test, and ship without losing the thread.'
            )}
          </p>
          <Link to='/about' className='robo-home-text-link'>
            {t('Learn more about RoboCoding')}
            <ArrowRight aria-hidden />
          </Link>
        </div>
        <div className='robo-home-workflow-list'>
          {workflowSteps.map(([title, description], index) => (
            <div key={title} className='robo-home-workflow-step'>
              <span className='robo-home-step-number'>0{index + 1}</span>
              <div>
                <h3>{t(title)}</h3>
                <p>{t(description)}</p>
              </div>
              <Check aria-hidden />
            </div>
          ))}
        </div>
      </section>

      <section className='robo-home-cta'>
        <div className='robo-home-cta-grid' aria-hidden='true' />
        <div className='robo-home-cta-content'>
          <div className='robo-home-cta-mark' aria-hidden='true'>
            R
          </div>
          <h2>{t('Make the next thing real.')}</h2>
          <p>{t('Start with the work in front of you.')}</p>
          <Button
            className='robo-home-cta-action'
            render={<Link to={consolePath} />}
          >
            {t('Enter RoboCoding')}
            <ArrowRight aria-hidden />
          </Button>
        </div>
      </section>
    </main>
  )
}
