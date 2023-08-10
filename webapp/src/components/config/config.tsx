import React, {useState, useCallback} from 'react';
import styled from 'styled-components';

import {ServiceData} from './service'
import ServiceForm from './service_form'
import Security, {SecurityConfig} from './security'

const AddAIServiceButton = styled.button`
    margin-bottom: 10px;
`

type Value = {
    services: ServiceData[],
    llmBackend: string,
    transcriptBackend: string,
    imageGeneratorBackend: string,
    enableLLMTrace: boolean,
    enableCallSummary: boolean,
    securityConfig: SecurityConfig
}

type Props = {
    id: string
    label: string
    helpText: React.ReactNode
    value: Value
    disabled: boolean
    config: any
    currentState: any
    license: any
    setByEnv: boolean
    onChange: (id: string, value: any) => void
    setSaveNeeded: () => void
}

const Config = (props: Props) => {
    const value = props.value || {services: [], llmBackend: '', transcriptBackend: '', imageGeneratorBackend: '', enableLLMTrace: false}
    const currentServices = value.services
    const securityConfig = value.securityConfig || {enableUserRestrictions: false, allowPrivateChannels: false, allowedTeamIds: '', onlyUsersOnTeam: ''}

    const addNewService = useCallback((e: React.MouseEvent) => {
        e.preventDefault()
        e.stopPropagation()
        const newService = {
            id: Math.random().toString(36).substring(2, 22),
            name: 'AI Engine',
            serviceName: 'openai',
            defaultModel: '',
            url: '',
            apiKey: '',
            username: '',
            password: '',
        }

        let counter = 1
        while (true) {
            let isNew = true
            for (const service of currentServices) {
                if (service.name == newService.name) {
                    isNew = false
                }
            }
            if (isNew) {
                break
            }
            newService.name = `AI Engine ${counter}`
            counter++
        }
        if (value.services.length === 0) {
            props.onChange(props.id, {...value, services: [...currentServices, newService], llmBackend: newService.name, transcriptBackend: newService.name, imageGeneratorBackend: newService.name})
        } else {
            props.onChange(props.id, {...value, services: [...currentServices, newService]})
        }
    }, [value, currentServices])

    return (
        <div>
            {currentServices.map((service, idx) => (
                <ServiceForm
                    key={idx}
                    service={service}
                    onDelete={(service) => {
                        const updatedServiceIdx = currentServices.indexOf(service)
                        if (updatedServiceIdx === -1) {
                            throw new Error('Service not found')
                        }
                        let newValue = value
                        if (currentServices.length > 1) {
                            if (value.llmBackend === service.name) {
                                newValue = {...newValue, llmBackend: value.services[0]?.name || ''}
                            }
                            if (value.imageGeneratorBackend === service.name) {
                                newValue = {...newValue, imageGeneratorBackend: value.services[0]?.name || ''}
                            }
                            if (value.transcriptBackend === service.name) {
                                newValue = {...newValue, transcriptBackend: value.services[0]?.name || ''}
                            }
                        } else {
                            newValue = {...newValue, llmBackend: '', transcriptBackend: '', imageGeneratorBackend: ''}
                        }
                        props.onChange(props.id, {...newValue, services: [...currentServices.slice(0, updatedServiceIdx), ...currentServices.slice(updatedServiceIdx + 1)]})
                        props.setSaveNeeded()
                    }}
                    onChange={(service) => {
                        const updatedServiceIdx = currentServices.findIndex((s) => service.id === s.id)
                        if (updatedServiceIdx === -1) {
                            throw new Error('Service not found')
                        }
                        let newValue = value
                        if (value.llmBackend === currentServices[updatedServiceIdx].name) {
                            newValue = {...newValue, llmBackend: service.name}
                        }
                        if (value.imageGeneratorBackend === currentServices[updatedServiceIdx].name) {
                            newValue = {...newValue, imageGeneratorBackend: service.name}
                        }
                        if (value.transcriptBackend === currentServices[updatedServiceIdx].name) {
                            newValue = {...newValue, transcriptBackend: service.name}
                        }
                        props.onChange(props.id, {...newValue, services: [...currentServices.slice(0, updatedServiceIdx), service, ...currentServices.slice(updatedServiceIdx + 1)]})
                        props.setSaveNeeded()
                    }}
                />
            ))}
            <AddAIServiceButton
                className='save-button btn btn-primary'
                onClick={addNewService}
            >
                Add AI Service
            </AddAIServiceButton>
            <div className='form-group'>
                <label
                    className='control-label col-sm-4'
                    htmlFor='ai-llm-backend'
                >
                    AI Large Language Model service
                </label>
                <div className='col-sm-8'>
                    <select
                        id='ai-llm-backend'
                        className={currentServices.length === 0 ? 'form-control disabled' : 'form-control'}
                        onChange={(e) => {
                            props.onChange(props.id, {...value, llmBackend: e.target.value})
                            props.setSaveNeeded()
                        }}
                        value={value.llmBackend}
                        disabled={currentServices.length === 0}
                    >
                        {currentServices.map((service) => (
                            <option
                                key={service.id}
                                value={service.name}
                            >
                                {service.name}
                            </option>
                        ))}
                    </select>
                    {currentServices.length === 0 && (
                        <div className="help-text">
                            <span>You need at least one AI services use this setting.</span>
                        </div>
                    )}
                </div>
            </div>
            <div className='form-group'>
                <label
                    className='control-label col-sm-4'
                    htmlFor='ai-image-generator'
                >
                    AI Image Generator service
                </label>
                <div className='col-sm-8'>
                    <select
                        id='ai-image-generator'
                        className={currentServices.length === 0 ? 'form-control disabled' : 'form-control'}
                        onChange={(e) => {
                            props.onChange(props.id, {...value, imageGeneratorBackend: e.target.value})
                            props.setSaveNeeded()
                        }}
                        value={value.imageGeneratorBackend}
                        disabled={currentServices.length === 0}
                    >
                        {currentServices.map((service) => (
                            <option
                                key={service.id}
                                value={service.name}
                            >
                                {service.name}
                            </option>
                        ))}
                    </select>
                    {currentServices.length === 0 && (
                        <div className="help-text">
                            <span>You need at least one AI services use this setting.</span>
                        </div>
                    )}
                </div>
            </div>
            <div className='form-group'>
                <label
                    className='control-label col-sm-4'
                    htmlFor='ai-transcript-backend'
                >
                    AI Audio/Video transcript service
                </label>
                <div className='col-sm-8'>
                    <select
                        id='ai-transcript-backend'
                        className={currentServices.length === 0 ? 'form-control disabled' : 'form-control'}
                        onChange={(e) => {
                            props.onChange(props.id, {...value, transcriptBackend: e.target.value})
                            props.setSaveNeeded()
                        }}
                        value={value.transcriptBackend}
                        disabled={currentServices.length === 0}
                    >
                        {currentServices.map((service) => (
                            <option
                                key={service.id}
                                value={service.name}
                            >
                                {service.name}
                            </option>
                        ))}
                    </select>
                    {currentServices.length === 0 && (
                        <div className="help-text">
                            <span>You need at least one AI services use this setting.</span>
                        </div>
                    )}
                </div>
            </div>

            <Security
                securityConfig={securityConfig}
                onChange={(securityConfig) => {
                    props.onChange(props.id, {...value, securityConfig})
                    props.setSaveNeeded()
                }}
            />

            <div className='form-group'>
                <label
                    className='control-label col-sm-4'
                >
                    Enable Automatic Call Sumary:
                </label>
                <div className="col-sm-8">
                    <label className="radio-inline">
                        <input
                            type="radio"
                            value="true"
                            checked={value.enableCallSummary}
                            onChange={() => props.onChange(props.id, {...value, enableCallSummary: true})}
                        />
                        <span>true</span>
                    </label>
                    <label className="radio-inline">
                        <input
                            type="radio"
                            value="false"
                            checked={!value.enableCallSummary}
                            onChange={() => props.onChange(props.id, {...value, enableCallSummary: false})}
                        />
                        <span>false</span>
                    </label>
                    <div className="help-text"><span>Automatically create a summary of any recorded call.</span></div>
                </div>
            </div>

            <div className='form-group'>
                <label
                    className='control-label col-sm-4'
                    htmlFor='ai-service-name'
                >
                    Enable LLM Trace:
                </label>
                <div className="col-sm-8">
                    <label className="radio-inline">
                        <input
                            type="radio"
                            value="true"
                            checked={value.enableLLMTrace}
                            onChange={() => props.onChange(props.id, {...value, enableLLMTrace: true})}
                        />
                        <span>true</span>
                    </label>
                    <label className="radio-inline">
                        <input
                            type="radio"
                            value="false"
                            checked={!value.enableLLMTrace}
                            onChange={() => props.onChange(props.id, {...value, enableLLMTrace: false})}
                        />
                        <span>false</span>
                    </label>
                    <div className="help-text"><span>Enable tracing of LLM requests. Outputs whole conversations to the logs.</span></div>
                </div>
            </div>
        </div>
    )
}
export default Config;
