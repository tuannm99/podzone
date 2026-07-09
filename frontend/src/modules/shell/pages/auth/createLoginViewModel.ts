import { createSignal } from 'solid-js'
import { useNavigate } from '@tanstack/solid-router'
import { login } from '@/services/auth'
import { createFormStore } from '@/solid/forms'

type LoginFormValues = {
    username: string
    password: string
}

export function createLoginViewModel() {
    const navigate = useNavigate()

    const form = createFormStore<LoginFormValues>({
        initialValues: { username: '', password: '' },
        validators: {
            username: [(v) => (!v?.trim() ? 'Username required' : undefined)],
            password: [(v) => (!v ? 'Password required' : undefined)],
        },
    })

    const [error, setError] = createSignal('')

    const submit = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!form.validate()) return
        setError('')
        form.setSubmitting(true)
        try {
            const { success, data } = await login({
                username: form.values.username.trim(),
                password: form.values.password,
            })
            if (!success) {
                setError(data?.message || 'Sign-in failed')
                return
            }
            void navigate({ to: '/admin', replace: true })
        } finally {
            form.setSubmitting(false)
        }
    }

    return { form, error, submit }
}
