import {useState, useEffect} from 'react';

import {useDispatch, useSelector} from 'react-redux';

import {GlobalState} from '@mattermost/types/store';

import {getAIBots} from '@/client';

import manifest from './manifest';
import {BotsHandler} from './redux';
import {ChannelAssistanceLevel, UserAssistanceLevel} from './components/system_console/bot';

export interface LLMBot {
    id: string;
    displayName: string;
    username: string;
    lastIconUpdate: number;
    dmChannelID: string;
    channelAssistanceLevel: ChannelAssistanceLevel;
    channelIDs: string[];
    userAssistanceLevel: UserAssistanceLevel;
    userIDs: string[];
}

const defaultBotLocalStorageKey = 'defaultBot';

export const useBotlist = () => {
    const bots = useSelector<GlobalState, LLMBot[] | null>((state: any) => state['plugins-' + manifest.id].bots);
    const [activeBot, setActiveBot] = useState<LLMBot | null>(null);
    const currentUserId = useSelector<GlobalState, string>((state) => state.entities.users.currentUserId);
    const dispatch = useDispatch();

    // Load bots
    useEffect(() => {
        const fetchBots = async () => {
            const fetchedBots = await getAIBots();
            if (!fetchedBots) {
                return;
            }

            dispatch({
                type: BotsHandler,
                bots: fetchedBots,
            });
        };
        if (!bots) {
            fetchBots();
        }
    }, [currentUserId, bots, dispatch]);

    useEffect(() => {
        const defaultActiveBotName = localStorage.getItem(defaultBotLocalStorageKey);
        setActiveBot(bots?.find((bot: LLMBot) => bot.username === defaultActiveBotName) || bots?.[0] || null);
    }, [bots]);

    useEffect(() => {
        if (!activeBot) {
            return;
        }
        localStorage.setItem(defaultBotLocalStorageKey, activeBot.username);
    }, [activeBot]);

    return {bots, activeBot, setActiveBot};
};

// useBotlistForChannel only shows bots the user is allowed to use in a specific channel. Also returns if bots were filtered for showing
// a sorry no bots message.
export const useBotlistForChannel = (channelId: string) => {
    const {bots, activeBot, setActiveBot} = useBotlist();
    const [filteredBots, setFilteredBots] = useState<LLMBot[]>([]);

    useEffect(() => {
        if (!bots) {
            return;
        }

        const filtered = bots.filter((bot: LLMBot) => {
            return bot.channelAssistanceLevel === ChannelAssistanceLevel.All ||
				(bot.channelAssistanceLevel === ChannelAssistanceLevel.Allow && bot.channelIDs.includes(channelId)) ||
				(bot.channelAssistanceLevel === ChannelAssistanceLevel.Block && !bot.channelIDs.includes(channelId));
        });

        setFilteredBots(filtered);
        if (!filtered.find((bot) => bot.username === activeBot?.username)) {
            setActiveBot(filtered[0] || null);
        }
    }, [bots, channelId, activeBot, setActiveBot]);

    const wasFiltered = bots && (filteredBots.length !== bots.length);

    return {bots: filteredBots, activeBot, setActiveBot, wasFiltered};
};
