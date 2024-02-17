import { test, expect } from '@playwright/test';

import RunContainer from 'helpers/plugincontainer';
import MattermostContainer from 'helpers/mmcontainer';
import {login} from 'helpers/mm';
import {openRHS} from 'helpers/ai-plugin';
import { RunOpenAIMocks } from 'helpers/openai-mock';

let mattermost: MattermostContainer;

test.beforeAll(async () => {
	mattermost = await RunContainer();
	await RunOpenAIMocks(mattermost.network)
});

test.afterAll(async () => {
	await mattermost.stop();
})

/*test('was installed', async ({ page }) => {
	const url = mattermost.url()
	await login(page, url, "regularuser", "regularuser");;
	await openRHS(page);
});*/


test('rhs bot interaction', async ({ page }) => {
	const url = mattermost.url()
	await login(page, url, "regularuser", "regularuser");;
	await openRHS(page);
	await page.getByTestId('reply_textbox').click();
	await page.getByTestId('reply_textbox').fill('Respond with "green"');
	await page.getByTestId('reply_textbox').press('Enter');
	await expect(page.getByText("Hello! How can I assist you today?")).toBeVisible();
})

