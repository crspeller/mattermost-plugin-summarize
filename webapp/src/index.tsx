import React from 'react';
import {Store, Action} from 'redux';

import {GlobalState} from '@mattermost/types/lib/store';

import {manifest} from '@/manifest';

import {LLMBotPost} from './components/llmbot_post';
import PostMenu from './components/post_menu';
import EditorMenu from './components/editor_menu';
import CodeMenu from './components/code_menu';
import IconThreadSummarization from './components/assets/icon_thread_summarization';
import IconReactForMe from './components/assets/icon_react_for_me';
import Config from './components/config/config';
import {doReaction, doSummarize, doTranscribe} from './client';
import {BotUsername} from './constants';
import PostEventListener from './websocket';
import {setupRedux} from './redux';

type WebappStore = Store<GlobalState, Action<Record<string, unknown>>>

const StreamingPostWebsocketEvent = 'custom_mattermost-ai_postupdate';

export default class Plugin {
    postEventListener: PostEventListener = new PostEventListener();

    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: any, store: WebappStore) {
        setupRedux(registry, store);

        registry.registerWebSocketEventHandler(StreamingPostWebsocketEvent, this.postEventListener.handlePostUpdateWebsockets);
        const LLMBotPostWithWebsockets = (props: any) => {
            return (
                <LLMBotPost
                    {...props}
                    websocketRegister={this.postEventListener.registerPostUpdateListener}
                    websocketUnregister={this.postEventListener.unregisterPostUpdateListener}
                />
            )
            ;
        };

        registry.registerPostTypeComponent('custom_llmbot', LLMBotPostWithWebsockets);
        if (registry.registerPostActionComponent) {
            registry.registerPostActionComponent(PostMenu);
        } else {
            registry.registerPostDropdownMenuAction(<><span className='icon'><IconThreadSummarization/></span>{'Summarize Thread'}</>, (postId: string) => {
                const state = store.getState();
                const team = state.entities.teams.teams[state.entities.teams.currentTeamId];
                window.WebappUtils.browserHistory.push('/' + team.name + '/messages/@' + BotUsername);
                doSummarize(postId);
            });
            registry.registerPostDropdownMenuAction(<><span className='icon'><IconThreadSummarization/></span>{'Summarize Meeting Audio'}</>, doTranscribe);
            registry.registerPostDropdownMenuAction(<><span className='icon'><IconReactForMe/></span>{'React for me'}</>, doReaction);
        }
        if (registry.registerPostEditorActionComponent) {
            registry.registerPostEditorActionComponent(EditorMenu);
        }

        registry.registerAdminConsoleCustomSetting('Config', Config);

        if (registry.registerCodeBlockActionComponent) {
            registry.registerCodeBlockActionComponent(CodeMenu);
        }
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
        WebappUtils: any
    }
}

window.registerPlugin(manifest.id, new Plugin());
