import { createSignal } from 'solid-js'
import { useNavigate } from '@tanstack/solid-router'
import { register } from '@podzone/shared/services/auth'
import { createFormStore } from '@podzone/shared/ui/forms'

type RegisterFormValues = {
    username: string
    email: string
    password: string
    confirmPassword: string
}

export function createRegisterViewModel() {
    const navigate = useNavigate()

    const form = createFormStore<RegisterFormValues>({
        initialValues: { username: '', email: '', password: '', confirmPassword: '' },
        validators: {
            username: [(v) => (!v?.trim() ? 'Username required' : undefined)],
            email: [(v) => (!v?.trim() ? 'Email required' : undefined)],
            password: [(v) => (!v ? 'Password required' : undefined)],
            confirmPassword: [(v, all) => (v !== all.password ? 'Passwords do not match' : undefined)],
        },
    })

    const [error, setError] = createSignal('')

    const submit = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!form.validate()) return
        setError('')
        form.setSubmitting(true)
        try {
            const { success, data } = await register({
                username: form.values.username.trim(),
                email: form.values.email.trim(),
                password: form.values.password,
            })
            if (!success) {
                setError(data?.message || 'Account setup failed')
                return
            }
            void navigate({ to: '/admin', replace: true })
        } finally {
            form.setSubmitting(false)
        }
    }

    return { form, error, submit }
}
