import { RouterProvider } from '@tanstack/solid-router'
import { render } from 'solid-js/web'
import { router } from './solid/app-router'
import './solid/global.css'

const root = document.getElementById('root')

if (!root) {
  throw new Error('Root element not found')
}

render(() => <RouterProvider router={router} />, root)
