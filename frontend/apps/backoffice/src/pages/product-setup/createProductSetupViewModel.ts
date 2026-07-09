import { useAuthContext } from '@podzone/shared/auth'
import { createEffect, createResource, createSignal, type Accessor } from 'solid-js'
import {
    createProductSetupDraft,
    getProductSetupSnapshot,
    promoteProductSetupCandidate,
    updateProductSetupCandidateStatus,
    type CatalogCandidate,
    type SetupDraft,
} from '@podzone/shared/services/productSetup'

import { createFormStore, required } from '@podzone/shared/ui/forms'
import { moneyValue, productSetupInitialValues } from './forms'

interface ProductSetupViewModelOptions {
    tenantID: Accessor<string>
    workspaceReady: Accessor<boolean>
}

export function createProductSetupViewModel(options: ProductSetupViewModelOptions) {
    const auth = useAuthContext()
    const form = createFormStore({
        initialValues: productSetupInitialValues,
        validators: {
            name: [required('Enter a product name.')],
            baseCost: [moneyValue('Use a valid amount, for example $8.20.')],
            retailPrice: [moneyValue('Use a valid amount, for example $24.00.')],
            variantColor: [required('Enter a primary color.')],
            variantSize: [required('Enter a primary size.')],
        },
    })
    const [message, setMessage] = createSignal('')
    const [error, setError] = createSignal('')
    const [promotingDraftID, setPromotingDraftID] = createSignal('')
    const [updatingCandidateID, setUpdatingCandidateID] = createSignal('')
    const [snapshot, { refetch: reload }] = createResource(
        () => (options.workspaceReady() ? options.tenantID() : undefined),
        async () => getProductSetupSnapshot()
    )
    const snapshotResult = () => snapshot.latest
    const readError = () => {
        const result = snapshotResult()
        if (error()) return error()
        return result && !result.success ? result.message : ''
    }
    const drafts = (): SetupDraft[] => {
        const result = snapshotResult()
        return result?.success ? result.data.drafts : []
    }
    const candidates = (): CatalogCandidate[] => {
        const result = snapshotResult()
        return result?.success ? result.data.candidates : []
    }

    const addDraft = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!form.validate()) return
        form.setSubmitting(true)
        setError('')
        setMessage('')
        try {
            const result = await createProductSetupDraft({
                name: form.values.name.trim(),
                partner: form.values.partner.trim(),
                baseCost: form.values.baseCost.trim(),
                retailPrice: form.values.retailPrice.trim(),
                status: form.values.status,
                notes: form.values.notes.trim(),
            })
            if (!result.success) {
                setError(result.message)
                return
            }
            setMessage(`Saved backend product setup draft for ${result.data.name}.`)
            form.reset()
            await reload()
        } finally {
            form.setSubmitting(false)
        }
    }

    const promoteToCandidate = async (draft: SetupDraft) => {
        setPromotingDraftID(draft.id)
        setError('')
        setMessage('')
        try {
            const result = await promoteProductSetupCandidate({
                draftId: draft.id,
                channel: form.values.channel,
                variantColor: form.values.variantColor.trim(),
                variantSize: form.values.variantSize.trim(),
                artworkChecklist: {
                    frontArtwork: form.values.hasFrontArtwork,
                    backArtwork: form.values.hasBackArtwork,
                    mockupReady: form.values.mockupReady,
                    printSpecChecked: form.values.printSpecChecked,
                },
                merchandisingNotes: form.values.notes.trim(),
            })
            if (!result.success) {
                setError(result.message)
                return
            }
            setMessage(`Promoted ${draft.name} into a backend catalog candidate.`)
            await reload()
        } finally {
            setPromotingDraftID('')
        }
    }

    const updateCandidateStatus = async (candidateID: string, nextStatus: string, successMessage: string) => {
        setUpdatingCandidateID(candidateID)
        setError('')
        setMessage('')
        try {
            const result = await updateProductSetupCandidateStatus(candidateID, nextStatus)
            if (!result.success) {
                setError(result.message)
                return
            }
            setMessage(successMessage)
            await reload()
        } finally {
            setUpdatingCandidateID('')
        }
    }

    createEffect(() => {
        auth.setActiveTenantId(options.tenantID())
    })

    createEffect(() => {
        if (snapshot.latest?.success) setError('')
    })

    return {
        form,
        message,
        error: readError,
        loading: () => snapshot.loading,
        promotingDraftID,
        updatingCandidateID,
        drafts,
        candidates,
        addDraft,
        promoteToCandidate,
        updateCandidateStatus,
        reload,
    }
}
