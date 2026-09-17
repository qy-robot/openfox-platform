/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

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
  ArrowDownToLine,
  ArrowRight,
  Bot,
  Check,
  Code2,
  Folder,
  Layers,
  Users,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

export function ProductHome(props: { isAuthenticated: boolean }) {
  const { t } = useTranslation()
  return (
    <main className='robo-home'>
      <section className='robo-hero mx-auto grid max-w-6xl items-center gap-12 px-6 pt-32 pb-20 lg:grid-cols-[1fr_1.05fr] lg:gap-14 lg:pt-40 lg:pb-28'>
        <div>
          <p className='text-primary mb-6 flex items-center gap-2 text-sm font-medium'>
            <Bot className='size-4' aria-hidden />
            {t('Your workspace for building robots')}
          </p>
          <h1 className='text-foreground text-[2.65rem] leading-[1.22] font-semibold tracking-tight sm:text-6xl'>
            {t('Turn your ideas')}
            <br />
            <span className='text-primary'>{t('into robots that work.')}</span>
          </h1>
          <p className='text-muted-foreground mt-7 max-w-md text-base leading-8'>
            {t(
              'Write code, use robotics skills, and manage AI services together in RoboCoding.'
            )}
          </p>
          <div className='mt-9 flex flex-wrap gap-3'>
            <Button
              className='h-12 gap-2 px-6'
              render={<Link to='/download' />}
            >
              <ArrowDownToLine aria-hidden />
              {t('Download desktop app')}
            </Button>
            <Button
              variant='outline'
              className='h-12 gap-2 px-6'
              render={
                <Link to={props.isAuthenticated ? '/dashboard' : '/sign-in'} />
              }
            >
              {t('Open console')}
              <ArrowRight aria-hidden />
            </Button>
          </div>
          <p className='text-muted-foreground mt-5 text-xs'>
            {t('One account for your desktop workspace and cloud services.')}
          </p>
        </div>
        <div className='robo-workspace border-border overflow-hidden rounded-2xl border shadow-xl shadow-slate-900/5'>
          <div className='bg-card border-border flex items-center justify-between border-b px-5 py-4'>
            <span className='flex items-center gap-2 text-sm font-semibold'>
              <Bot className='text-primary size-5' aria-hidden />
              RoboCoding
            </span>
            <span className='text-muted-foreground text-xs'>
              {t('Workspace preview')}
            </span>
          </div>
          <div className='bg-card grid min-h-80 grid-cols-[88px_1fr] sm:grid-cols-[112px_1fr]'>
            <div className='bg-muted/40 border-border space-y-5 border-r px-3 py-6 text-xs'>
              <div className='text-primary flex items-center gap-2'>
                <Code2 className='size-4' aria-hidden />
                {t('Build')}
              </div>
              <div className='text-muted-foreground flex items-center gap-2'>
                <Layers className='size-4' aria-hidden />
                {t('Skills')}
              </div>
              <div className='text-muted-foreground flex items-center gap-2'>
                <Folder className='size-4' aria-hidden />
                {t('Files')}
              </div>
            </div>
            <div className='min-w-0 p-5 sm:p-6'>
              <p className='text-muted-foreground mb-5 text-xs'>
                robot-workspace /
              </p>
              <div className='bg-accent text-accent-foreground rounded-xl rounded-tr-sm px-4 py-3 text-sm leading-6'>
                {t('Help me understand this robot project.')}
              </div>
              <div className='mt-6 flex items-start gap-3'>
                <Bot
                  className='text-primary mt-0.5 size-5 shrink-0'
                  aria-hidden
                />
                <div>
                  <p className='text-sm font-medium'>
                    {t('Start with a clear plan.')}
                  </p>
                  <p className='text-muted-foreground mt-2 text-xs leading-6'>
                    {t(
                      'Explore the files, choose a skill, then review each change.'
                    )}
                  </p>
                </div>
              </div>
              <div className='border-border mt-6 space-y-3 rounded-lg border p-4'>
                {['Project context', 'Robotics skills', 'Code and changes'].map(
                  (label) => (
                    <div
                      key={label}
                      className='text-muted-foreground flex items-center gap-2 text-xs'
                    >
                      <Check className='text-primary size-3.5' aria-hidden />
                      {t(label)}
                    </div>
                  )
                )}
              </div>
            </div>
          </div>
          <div className='bg-muted/40 border-border text-muted-foreground border-t px-5 py-3 text-xs'>
            {t(
              'An illustration of the workflow. Your projects stay in your workspace.'
            )}
          </div>
        </div>
      </section>
      <section className='border-border bg-card/70 border-y'>
        <div className='mx-auto grid max-w-6xl gap-8 px-6 py-9 md:grid-cols-3'>
          {[
            {
              icon: Code2,
              title: 'From a task to working code',
              description:
                'Keep conversations, project files, and changes in one workspace.',
            },
            {
              icon: Layers,
              title: 'Skills for robotics work',
              description:
                'Find a skill for the task and add it when you need it.',
            },
            {
              icon: Users,
              title: 'One place for your team',
              description:
                'Manage members, shared balance, and usage from the console.',
            },
          ].map(({ icon: FeatureIcon, title, description }) => {
            return (
              <div key={title} className='flex gap-4'>
                <FeatureIcon
                  className='text-primary mt-1 size-5 shrink-0'
                  aria-hidden
                />
                <div>
                  <h2 className='text-sm font-semibold'>{t(title)}</h2>
                  <p className='text-muted-foreground mt-2 text-sm leading-6'>
                    {t(description)}
                  </p>
                </div>
              </div>
            )
          })}
        </div>
      </section>
      <section className='mx-auto grid max-w-6xl gap-8 px-6 py-20 md:grid-cols-[1fr_1.15fr] md:gap-20'>
        <div>
          <p className='text-primary mb-4 text-sm font-medium'>
            {t('Built around the way you work')}
          </p>
          <h2 className='text-3xl leading-snug font-semibold tracking-tight'>
            {t('Learn by building.')}
            <br />
            {t('Move forward with every task.')}
          </h2>
        </div>
        <div className='space-y-7'>
          {[
            [
              'Bring your project',
              'Open a local project and describe what you want to build.',
            ],
            [
              'Choose your tools',
              'Use official AI services or connect your own model service.',
            ],
            [
              'Stay in control',
              'Review changes in the app and track account usage in the console.',
            ],
          ].map(([title, description], index) => (
            <div key={title} className='flex gap-5'>
              <span className='text-primary bg-accent flex size-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold'>
                {index + 1}
              </span>
              <div>
                <h3 className='font-medium'>{t(title)}</h3>
                <p className='text-muted-foreground mt-2 text-sm leading-6'>
                  {t(description)}
                </p>
              </div>
            </div>
          ))}
        </div>
      </section>
    </main>
  )
}
