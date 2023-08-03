( function()
{
	const container = document.querySelector( '#js-realtime-log' );
	const usersOnline = document.querySelector( '#js-realtime-users' );

	container.innerHTML = '';

	const GetTime = function()
	{
		const date = new Date();
		//let hh = date.getUTCHours();
		let hh = date.getHours();
		//let mm = date.getUTCMinutes();
		let mm = date.getMinutes();
		let ss = date.getSeconds();

		if( hh < 10 )
		{
			hh = '0' + hh;
		}
		if( mm < 10 )
		{
			mm = '0' + mm;
		}
		if( ss < 10 )
		{
			ss = '0' + ss;
		}

		return hh + ':' + mm + ':' + ss;
	};

	const AddToLog = function( text )
	{
		const element = document.createElement( 'div' );
		element.className = 'line';

		const time = document.createElement( 'span' );
		time.textContent = GetTime();
		time.className = 'time';
		element.appendChild( time );

		element.insertAdjacentHTML( 'beforeend', text );

		container.prepend( element );

		// Keep only 1000 log lines
		for( let i = container.childNodes.length - 1; i > 1000; i-- )
		{
			container.childNodes[ i ].remove();
		}
	};

	AddToLog( '初始化中… 由 <a href="https://github.com/xPaw/SteamWebPipes">SteamWebPipes</a> 驱动。' );

	let reconnectAttempts = 1;

	function generateInterval( k )
	{
		const maxInterval = Math.pow( 2, k ) * 1000;

		return 5000 * Math.random() + maxInterval;
	}

	function createWebSocket()
	{
		const connection = new WebSocket( 'ws://localhost:8181', [ 'steam-pics' ] );

		connection.onopen = function()
		{
			reconnectAttempts = 1;

			AddToLog( '连接已打开，正在等待…' );
		};

		connection.onclose = function()
		{
			usersOnline.textContent = '?';

			if( reconnectAttempts > 7 )
			{
				AddToLog( '无法与后台服务器建立初始化连接，将不再重试。请手动刷新页面。' );

				return;
			}

			const time = generateInterval( reconnectAttempts );

			AddToLog( '连接已中断，将于 ' + Math.round( time / 1000 ) + ' 秒后重试…' );

			setTimeout( function()
			{
				reconnectAttempts++;
				createWebSocket();
			}, time );
		};

		connection.onmessage = function( e )
		{
			const data = JSON.parse( e.data );

			switch( data.Type )
			{
				case 'UsersOnline':
				{
					usersOnline.textContent = data.Users;
					break;
				}

				case 'LogOn':
				{
					AddToLog( 'Bot 已登录至 Steam，正在检查新的 Changelist…' );
					break;
				}

				case 'LogOff':
				{
					AddToLog( 'Bot 已从 Steam 登出，即将重试连接…' );
					break;
				}

				case 'Changelist':
				{
					let str = 'Changelist <a href="https://steamdb.info/changelist/' + data.ChangeNumber + '/" target="_blank" class="muted" rel="nofollow">#' + data.ChangeNumber + '</a>';

					let list = [];

					for( let [ appid, value ] of Object.entries( data.Apps ) )
					{
						appid = +appid;
						value = value.replace( /&/g, '&amp;' ).replace( /</g, '&lt;' ).replace( />/g, '&gt;' );
						list.push( `<a href="https://steamdb.info/app/${appid}/history/" target="_blank" rel="noopener">${value}</a>` );
					}

					if( list.length )
					{
						str += ' — 应用 (' + list.length + '): ' + list.join( ', ' );
					}

					list = [];

					for( let [ subid, value ] of Object.entries( data.Packages ) )
					{
						subid = +subid;
						value = value.replace( /&/g, '&amp;' ).replace( /</g, '&lt;' ).replace( />/g, '&gt;' );
						list.push( `<a href="https://steamdb.info/sub/${subid}/history/" target="_blank" rel="noopener">${value}</a>` );
					}

					if( list.length )
					{
						str += ' — 捆绑包 (' + list.length + '): ' + list.join( ', ' );
					}

					AddToLog( str );

					break;
				}

				default:
				{
					AddToLog( '接收到未知事件 ' + data.Type );
				}
			}
		};
	}

	createWebSocket();
}() );
