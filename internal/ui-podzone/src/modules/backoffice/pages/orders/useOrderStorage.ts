import { createSignal } from 'solid-js';
import type {
  BulkDraft,
  SavedBulkTemplate,
  SavedQueuePreset,
  ShipmentSlaMode,
} from './board-context';
import {
  bulkTemplateStorageKey,
  queuePresetStorageKey,
  resolveShipmentSla,
  toLocalDateTimeValue,
} from './presentation';

type OrderStorageParams = {
  tenantId: () => string;
  storeId: () => string;
  activeQueueView: () => SavedQueuePreset['queueView'];
  setActiveQueueView: (value: SavedQueuePreset['queueView']) => void;
  activeQueueSort: () => SavedQueuePreset['queueSort'];
  setActiveQueueSort: (value: SavedQueuePreset['queueSort']) => void;
  operatorLens: () => string;
  setOperatorLens: (value: string) => void;
  setMessage: (value: string) => void;
};

export function useOrderStorage(params: OrderStorageParams) {
  const [savedPresets, setSavedPresets] = createSignal<SavedQueuePreset[]>([]);
  const [presetName, setPresetName] = createSignal('');
  const [savedBulkTemplates, setSavedBulkTemplates] = createSignal<
    SavedBulkTemplate[]
  >([]);
  const [bulkTemplateName, setBulkTemplateName] = createSignal('');
  const [selectedOrderIDs, setSelectedOrderIDs] = createSignal<string[]>([]);
  const [bulkDraft, setBulkDraft] = createSignal<BulkDraft>({
    operatorAssignee: '',
    shipmentSlaDueAt: '',
    shipmentSlaMode: '',
    settlementStatus: '',
  });

  const loadSavedPresets = () => {
    const raw = window.localStorage.getItem(
      queuePresetStorageKey(params.tenantId(), params.storeId())
    );
    if (!raw) {
      setSavedPresets([]);
      return;
    }
    try {
      const parsed = JSON.parse(raw) as SavedQueuePreset[];
      setSavedPresets(Array.isArray(parsed) ? parsed : []);
    } catch {
      setSavedPresets([]);
    }
  };

  const persistSavedPresets = (next: SavedQueuePreset[]) => {
    window.localStorage.setItem(
      queuePresetStorageKey(params.tenantId(), params.storeId()),
      JSON.stringify(next)
    );
    setSavedPresets(next);
  };

  const loadSavedBulkTemplates = () => {
    const raw = window.localStorage.getItem(
      bulkTemplateStorageKey(params.tenantId(), params.storeId())
    );
    if (!raw) {
      setSavedBulkTemplates([]);
      return;
    }
    try {
      const parsed = JSON.parse(raw) as SavedBulkTemplate[];
      setSavedBulkTemplates(Array.isArray(parsed) ? parsed : []);
    } catch {
      setSavedBulkTemplates([]);
    }
  };

  const persistSavedBulkTemplates = (next: SavedBulkTemplate[]) => {
    window.localStorage.setItem(
      bulkTemplateStorageKey(params.tenantId(), params.storeId()),
      JSON.stringify(next)
    );
    setSavedBulkTemplates(next);
  };

  const saveQueuePreset = () => {
    const name = presetName().trim();
    if (!name) {
      params.setMessage('Enter a preset name first.');
      return;
    }
    const nextPreset: SavedQueuePreset = {
      name,
      queueView: params.activeQueueView(),
      queueSort: params.activeQueueSort(),
      operatorLens: params.operatorLens().trim(),
    };
    const deduped = savedPresets().filter((preset) => preset.name !== name);
    persistSavedPresets([nextPreset, ...deduped]);
    setPresetName('');
    params.setMessage(`Saved queue preset ${name}.`);
  };

  const applyQueuePreset = (preset: SavedQueuePreset) => {
    params.setActiveQueueView(preset.queueView);
    params.setActiveQueueSort(preset.queueSort);
    params.setOperatorLens(preset.operatorLens);
    params.setMessage(`Applied queue preset ${preset.name}.`);
  };

  const deleteQueuePreset = (name: string) => {
    persistSavedPresets(
      savedPresets().filter((preset) => preset.name !== name)
    );
    params.setMessage(`Deleted queue preset ${name}.`);
  };

  const saveBulkTemplate = () => {
    const name = bulkTemplateName().trim();
    if (!name) {
      params.setMessage('Enter a bulk template name first.');
      return;
    }
    const draft = bulkDraft();
    const nextTemplate: SavedBulkTemplate = {
      name,
      operatorAssignee: draft.operatorAssignee.trim(),
      shipmentSlaMode: draft.shipmentSlaMode,
      settlementStatus: draft.settlementStatus.trim(),
    };
    const deduped = savedBulkTemplates().filter((item) => item.name !== name);
    persistSavedBulkTemplates([nextTemplate, ...deduped]);
    setBulkTemplateName('');
    params.setMessage(`Saved bulk template ${name}.`);
  };

  const applyBulkTemplate = (template: SavedBulkTemplate) => {
    setBulkDraft({
      operatorAssignee: template.operatorAssignee,
      shipmentSlaMode: template.shipmentSlaMode,
      shipmentSlaDueAt: template.shipmentSlaMode
        ? toLocalDateTimeValue(resolveShipmentSla(template.shipmentSlaMode))
        : '',
      settlementStatus: template.settlementStatus,
    });
    params.setMessage(`Loaded bulk template ${template.name}.`);
  };

  const deleteBulkTemplate = (name: string) => {
    persistSavedBulkTemplates(
      savedBulkTemplates().filter((template) => template.name !== name)
    );
    params.setMessage(`Deleted bulk template ${name}.`);
  };

  const toggleOrderSelection = (orderID: string, checked: boolean) => {
    setSelectedOrderIDs((current) => {
      if (checked) {
        return current.includes(orderID) ? current : [...current, orderID];
      }
      return current.filter((id) => id !== orderID);
    });
  };

  const clearSelectedOrders = () => setSelectedOrderIDs([]);
  const isSelected = (orderID: string) => selectedOrderIDs().includes(orderID);

  const applyRelativeShipmentSla = (mode: ShipmentSlaMode) => {
    setBulkDraft((current) => ({
      ...current,
      shipmentSlaMode: mode,
      shipmentSlaDueAt: mode
        ? toLocalDateTimeValue(resolveShipmentSla(mode))
        : current.shipmentSlaDueAt,
    }));
  };

  return {
    savedPresets,
    presetName,
    setPresetName,
    savedBulkTemplates,
    bulkTemplateName,
    setBulkTemplateName,
    selectedOrderIDs,
    setSelectedOrderIDs,
    bulkDraft,
    setBulkDraft,
    loadSavedPresets,
    loadSavedBulkTemplates,
    saveQueuePreset,
    applyQueuePreset,
    deleteQueuePreset,
    saveBulkTemplate,
    applyBulkTemplate,
    deleteBulkTemplate,
    toggleOrderSelection,
    clearSelectedOrders,
    isSelected,
    applyRelativeShipmentSla,
  };
}
