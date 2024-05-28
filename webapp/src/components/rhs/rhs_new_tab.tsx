import React, {useState} from 'react';
import styled from 'styled-components';

import {
    FormatListNumberedIcon,
    LightbulbOutlineIcon,
    PlaylistCheckIcon,
} from '@mattermost/compass-icons/components';

import {useDispatch} from 'react-redux';

import RHSImage from '../assets/rhs_image';

import {createPost} from '@/client';

import {Button, RHSPaddingContainer, RHSText, RHSTitle} from './common';

const CreatePost = (window as any).Components.CreatePost;

const CreatePostContainer = styled.div`
	.custom-textarea {
		padding-top: 13px;
		padding-bottom: 13px;
		passing-left: 16px;
	}
    .AdvancedTextEditor {
        padding: 0px;
    }
`;

const OptionButton = styled(Button)`
    color: rgb(var(--link-color-rgb));
    background-color: rgba(var(--button-bg-rgb), 0.08);
    svg {
        fill: rgb(var(--link-color-rgb));
    }
    &:hover {
        background-color: rgba(var(--button-bg-rgb), 0.12);
	}
	font-weight: 600;
	line-height: 16px;
	font-size: 12px;
`;

const QuestionOptions = styled.div`
    display: flex;
	gap: 8px;
	margin-top: 16px;
	margin-bottom: 24px;
    flex-wrap: wrap;
`;

const PlusMinus = styled.i`
    width: 14px;
    font-size: 14px;
    font-weight: 400;
    margin-right: 4px;
`;

type Props = {
    botChannelId: string
    selectPost: (postId: string) => void
    setCurrentTab: (tab: string) => void
}

const setEditorText = (text: string) => {
    const replyBox = document.getElementById('reply_textbox');
    if (replyBox) {
        replyBox.innerHTML = text;
        replyBox.dispatchEvent(new Event('input', {bubbles: true}));
        replyBox.focus();
    }
};

const addBrainstormingIdeas = () => {
    setEditorText('Brainstorm ideas about ');
};

const addMeetingAgenda = () => {
    setEditorText('Write a meeting agenda about ');
};

const addToDoList = () => {
    setEditorText('Write a todo list about ');
};

const addProsAndCons = () => {
    setEditorText('Write a pros and cons list about ');
};

const RHSNewTab = ({botChannelId, selectPost, setCurrentTab}: Props) => {
    const dispatch = useDispatch();
    const [draft, updateDraft] = useState<any>(null);
    return (
        <RHSPaddingContainer>
            <RHSImage/>
            <RHSTitle>{'Ask AI Copilot anything'}</RHSTitle>
            <RHSText>{'The AI Copilot is here to help. Choose from the prompts below or write your own.'}</RHSText>
            <QuestionOptions>
                <OptionButton onClick={addBrainstormingIdeas}><LightbulbOutlineIcon/>{'Brainstorm ideas'}</OptionButton>
                <OptionButton onClick={addMeetingAgenda}><FormatListNumberedIcon/>{'Meeting agenda'}</OptionButton>
                <OptionButton onClick={addProsAndCons}><PlusMinus className='icon'>{'±'}</PlusMinus>{'Pros and Cons'}</OptionButton>
                <OptionButton onClick={addToDoList}><PlaylistCheckIcon/>{'To-do list'}</OptionButton>
            </QuestionOptions>
            <CreatePostContainer>
                <CreatePost
                    data-testid='rhs-new-tab-create-post'
                    channelId={botChannelId}
                    placeholder={'Ask AI Copilot anything...'}
                    rootId={'ai_copilot'}
                    onSubmit={async (p: any) => {
                        const post = {...p};
                        post.channel_id = botChannelId || '';
                        post.props = {};
                        post.uploadsInProgress = [];
                        post.file_ids = p.fileInfos.map((f: any) => f.id);
                        const created = await createPost(post);
                        selectPost(created.id);
                        setCurrentTab('thread');
                        dispatch({
                            type: 'SET_GLOBAL_ITEM',
                            data: {
                                name: 'comment_draft_ai_copilot',
                                value: {message: '', fileInfos: [], uploadsInProgress: []},
                            },
                        });
                    }}
                    draft={draft}
                    onUpdateCommentDraft={(newDraft: any) => {
                        updateDraft(newDraft);
                        const timestamp = new Date().getTime();
                        newDraft.updateAt = timestamp;
                        newDraft.createAt = newDraft.createAt || timestamp;
                        dispatch({
                            type: 'SET_GLOBAL_ITEM',
                            data: {
                                name: 'comment_draft_ai_copilot',
                                value: newDraft,
                            },
                        });
                    }}
                />
            </CreatePostContainer>
        </RHSPaddingContainer>
    );
};

export default React.memo(RHSNewTab);
