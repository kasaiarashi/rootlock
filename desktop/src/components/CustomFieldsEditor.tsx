import { CustomField } from '../types';
import './CustomFieldsEditor.css';

interface CustomFieldsEditorProps {
  fields: CustomField[];
  onChange: (fields: CustomField[]) => void;
  disabled?: boolean;
}

export default function CustomFieldsEditor({ fields, onChange, disabled }: CustomFieldsEditorProps) {
  const addField = () => {
    const newField: CustomField = {
      id: crypto.randomUUID(),
      label: '',
      value: '',
      type: 'text',
      hidden: false
    };
    onChange([...fields, newField]);
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
        <button
          type="button"
          onClick={addField}
          className="btn-add-field"
          disabled={disabled}
        >
          + Add Field
        </button>
      </div>

      {fields.length > 0 && (
        <div className="custom-fields-list">
          {fields.map((field) => (
            <div key={field.id} className="custom-field-item">
              <div className="field-row-top">
                <div className="field-label-group">
                  <label className="field-label-text">Field Name</label>
                  <input
                    type="text"
                    value={field.label}
                    onChange={(e) => updateField(field.id, { label: e.target.value })}
                    placeholder="e.g., Security Question"
                    className="field-label-input"
                    disabled={disabled}
                  />
                </div>
                
                <div className="field-type-group">
                  <label className="field-label-text">Type</label>
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
                </div>

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

              <div className="field-value-group">
                <label className="field-label-text">Value</label>
                <input
                  type={field.hidden ? 'password' : 'text'}
                  value={field.value}
                  onChange={(e) => updateField(field.id, { value: e.target.value })}
                  placeholder="Enter value..."
                  className="field-value-input"
                  disabled={disabled}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
