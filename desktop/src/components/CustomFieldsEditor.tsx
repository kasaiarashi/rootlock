import { useState } from 'react';
import { CustomField } from '../types';
import './CustomFieldsEditor.css';

interface CustomFieldsEditorProps {
  fields: CustomField[];
  onChange: (fields: CustomField[]) => void;
  disabled?: boolean;
}

export default function CustomFieldsEditor({ fields, onChange, disabled }: CustomFieldsEditorProps) {
  const [showAddField, setShowAddField] = useState(false);

  const addField = () => {
    const newField: CustomField = {
      id: crypto.randomUUID(),
      label: '',
      value: '',
      type: 'text',
      hidden: false
    };
    onChange([...fields, newField]);
    setShowAddField(false);
  };

  const updateField = (id: string, updates: Partial<CustomField>) => {
    onChange(fields.map(f => f.id === id ? { ...f, ...updates } : f));
  };

  const removeField = (id: string) => {
    onChange(fields.filter(f => f.id !== id));
  };

  return (
    <div className="custom-fields-editor">
      <div className="custom-fields-header">
        <label>Custom Fields</label>
        {!showAddField && (
          <button
            type="button"
            onClick={() => setShowAddField(true)}
            className="btn-add-field"
            disabled={disabled}
          >
            + Add Field
          </button>
        )}
      </div>

      {fields.length > 0 && (
        <div className="custom-fields-list">
          {fields.map((field) => (
            <div key={field.id} className="custom-field-item">
              <div className="field-inputs">
                <input
                  type="text"
                  value={field.label}
                  onChange={(e) => updateField(field.id, { label: e.target.value })}
                  placeholder="Field name (e.g., Security Question)"
                  className="field-label-input"
                  disabled={disabled}
                />
                
                <select
                  value={field.type}
                  onChange={(e) => updateField(field.id, { 
                    type: e.target.value as CustomField['type'],
                    hidden: e.target.value === 'password'
                  })}
                  className="field-type-select"
                  disabled={disabled}
                >
                  <option value="text">Text</option>
                  <option value="password">Password</option>
                  <option value="email">Email</option>
                  <option value="url">URL</option>
                  <option value="number">Number</option>
                </select>

                <input
                  type={field.hidden ? 'password' : 'text'}
                  value={field.value}
                  onChange={(e) => updateField(field.id, { value: e.target.value })}
                  placeholder="Value"
                  className="field-value-input"
                  disabled={disabled}
                />

                <button
                  type="button"
                  onClick={() => removeField(field.id)}
                  className="btn-remove-field"
                  disabled={disabled}
                  title="Remove field"
                >
                  🗑️
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showAddField && (
        <div className="add-field-prompt">
          <p>Click to add a new custom field</p>
          <div className="add-field-actions">
            <button
              type="button"
              onClick={addField}
              className="btn-confirm-add"
            >
              Add Field
            </button>
            <button
              type="button"
              onClick={() => setShowAddField(false)}
              className="btn-cancel-add"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
